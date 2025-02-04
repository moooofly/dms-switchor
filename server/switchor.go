package server

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gomodule/redigo/redis"
	"github.com/moooofly/dms-switchor/pkg/parser"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	radar "github.com/moooofly/radar-go-client"

	pb "github.com/moooofly/dms-switchor/proto"
)

type dmsMode int32

const (
	single_point dmsMode = 1
	master_slave dmsMode = 2
	cluster      dmsMode = 3
)

var str2dmsMode = map[string]dmsMode{
	"single-point": single_point,
	"master-slave": master_slave,
	"cluster":      cluster,
}

var dmsMode2str = map[dmsMode]string{
	single_point: "single-point",
	master_slave: "master-slave",
	cluster:      "cluster",
}

// used by Redis, MySQL and so on
type appRole int32

const (
	appIsMaster    appRole = 0
	appIsSlave     appRole = 1
	appIsCandidate appRole = 2
)

var appRole2str = map[appRole]string{
	appIsMaster:    "master",
	appIsSlave:     "slave",
	appIsCandidate: "candidate",
}

var appRole2electorRole = map[appRole]pb.EnumRole{
	appIsMaster:    pb.EnumRole_Leader,
	appIsSlave:     pb.EnumRole_Follower,
	appIsCandidate: pb.EnumRole_Candidate,
}

var electorRole2appRole = map[pb.EnumRole]appRole{
	pb.EnumRole_Leader:    appIsMaster,
	pb.EnumRole_Follower:  appIsSlave,
	pb.EnumRole_Candidate: appIsCandidate,
}

const (
	mysql_master_repl_result_file_index = int32(0)
	mysql_master_repl_result_pos_index  = int32(1)
	mysql_slave_repl_result_file_index  = int32(5)
	mysql_slave_repl_result_pos_index   = int32(21)
)

// FIXME: appName seems useless, just for log info
func is_role_match(appName string, ar appRole, er pb.EnumRole) bool {
	if appRole2electorRole[ar] == er {
		logrus.Infof("[switchor] <%s> appRole(%s) and electorRole(%s) --> [Match]", appName, appRole2str[ar], er)
		return true
	} else {
		logrus.Infof("[switchor] <%s> appRole(%s) and electorRole(%s) --> [Mismatch]", appName, appRole2str[ar], er)
		return false
	}
}

// ------------
//  Operators
// ------------

type Operator interface {
	check_and_switch(electorRole pb.EnumRole)
}

type modbOperator struct {
	sw *Switchor

	DomainMoid      string
	MachineRoomMoid string // equal to resourceMoid
	GroupMoid       string
	ServerMoid      string // moid
}

func newModbOperator(sw *Switchor, dm, rm, gm, sm string) *modbOperator {
	return &modbOperator{
		sw: sw,

		DomainMoid:      dm,
		MachineRoomMoid: rm, // equal to resourceMoid
		GroupMoid:       gm,
		ServerMoid:      sm,
	}
}

func (op *modbOperator) check_and_switch(electorRole pb.EnumRole) {
	match, err := op.check_modb_role_match(electorRole)
	if err != nil {
		logrus.Errorf("[switchor] check_modb_role_match() failed (err: %v), need to try again", err)
	} else {
		if !match {
			op.switch_modb_role(electorRole)
		}
	}
}

func (op *modbOperator) check_modb_role_match(electorRole pb.EnumRole) (bool, error) {

	a := radar.ReqArgs{
		DomainMoid:   op.DomainMoid,
		ResourceMoid: op.MachineRoomMoid,
		GroupMoid:    op.GroupMoid,
		ServerMoid:   op.ServerMoid,
		ServerName:   "modb",
	}

	// FIXME: 需要对 radarCli 是否可用做判定，并触发重连
	if op.sw.radarCli == nil {
		return false, fmt.Errorf("radarCli is nil now, wait some time")
	}

	rspGroup, err := op.sw.radarCli.GetAppControl([]radar.ReqArgs{a})
	if err != nil {
		logrus.Errorf("[switchor] GetAppControl failed: %v", err)
		return false, err
	}

	if len(rspGroup) != 1 {
		// NOTE: just warning here, no need to fail
		logrus.Warnf("[switchor] GetAppControl should return only 1 response, but got %v", rspGroup)
		//return
	}

	var curAppRole appRole
	switch rspGroup[0].Control {
	case radar.AppEnabled:
		curAppRole = appIsMaster
	case radar.AppDisabled:
		curAppRole = appIsSlave
	default:
		logrus.Errorf("[switchor] get wrong control value '%d' from radar server", rspGroup[0].Control)
	}
	return is_role_match("modb", curAppRole, electorRole), nil
}

func (op *modbOperator) switch_modb_role(electorRole pb.EnumRole) {

	// FIXME: 是否有必要对 radarCli 是否可用做判定，并主动触发重连？
	// radar server 的重连已经通过 heartbeat 后台进行
	if op.sw.radarCli == nil {
		logrus.Warnf("[switchor] radarCli is <nil>, wait and try again.")
		return
	}

	var oper string
	if electorRole == pb.EnumRole_Leader {
		oper = "1" // enabled
	} else {
		oper = "0" // disabled
	}

	ok, err := op.sw.radarCli.SetAppControl(oper, op.DomainMoid, op.MachineRoomMoid, op.GroupMoid, op.ServerMoid, "modb")
	if err != nil {
		logrus.Errorf("[switchor] radarCli.SetAppControl() failed: %v", err)
		return
	}

	// FIXME: 实现的丑陋
	if ok {
		logrus.Info("[switchor] switch modb's 'role to '%v' success", appRole2str[electorRole2appRole[electorRole]])
	} else {
		logrus.Warnf("[switchor] switch modb's 'role to '%v' failed", appRole2str[electorRole2appRole[electorRole]])
	}
}

type redisOperator struct {
	sw *Switchor

	LocalAddr      string
	LocalPassword  string
	RemoteAddr     string
	RemotePassword string

	c redis.Conn
}

func newRedisOperator(sw *Switchor, la, lp, ra, rp string) *redisOperator {
	return &redisOperator{
		sw: sw,

		LocalAddr:      la,
		LocalPassword:  lp,
		RemoteAddr:     ra,
		RemotePassword: rp,
	}
}

func (op *redisOperator) check_and_switch(electorRole pb.EnumRole) {

	match, err := op.check_redis_role_match(electorRole)
	if err != nil {
		logrus.Errorf("[switchor] check_redis_role_match() failed (err: %v), need to try again", err)
	} else {
		if !match {
			op.switch_redis_role(electorRole)
		}
	}
}

func isMaster(conn redis.Conn) bool {
	role, err := getRole(conn)
	if err != nil || role != "master" {
		return false
	}
	return true
}

// getRole is supplied to query an instance (master or slave) for its role.
// It attempts to use the ROLE command introduced in redis 2.8.12.
func getRole(c redis.Conn) (string, error) {
	res, err := c.Do("ROLE")
	if err != nil {
		return "", err
	}
	rres, ok := res.([]interface{})
	if ok {
		return redis.String(rres[0], nil)
	}
	return "", errors.New("redigo: can not transform ROLE reply to string")
}

// addr should be
// 1. "no one"
// 2. "<ip> <port>"
func slaveof(c redis.Conn, addr []string) (string, error) {
	res, err := c.Do("slaveof", addr[0], addr[1])
	if err != nil {
		return "", err
	}
	rres, ok := res.(string)
	if ok {
		return redis.String(rres, nil)
	}
	return "", errors.New("redigo: can not transform SLAVEOF reply to string")
}

func (op *redisOperator) check_redis_role_match(electorRole pb.EnumRole) (bool, error) {
	var curAppRole appRole
	if isMaster(op.c) {
		curAppRole = appIsMaster
	} else {
		curAppRole = appIsSlave
	}
	logrus.Debugf("[switchor] <redisOperator> obtain redis role ==> %v", appRole2str[curAppRole])

	return is_role_match("redis", curAppRole, electorRole), nil
}

func (op *redisOperator) switch_redis_role(electorRole pb.EnumRole) {

	switch electorRole {
	case pb.EnumRole_Leader:
		if res, err := slaveof(op.c, []string{"no", "one"}); err != nil {
			logrus.Warnf("[switchor] exec 'SLAVEOF NO ONE' failed, %v", err)
		} else {
			logrus.Infof("[switchor] exec 'SLAVEOF NO ONE' success, %s", res)
		}
	case pb.EnumRole_Follower:
		addr := strings.Split(op.RemoteAddr, ":")
		if res, err := slaveof(op.c, addr); err != nil {
			logrus.Warnf("[switchor] exec 'SLAVEOF %s' failed, %v", strings.Join(addr, " "), err)
		} else {
			logrus.Infof("[switchor] exec 'SLAVEOF %s' success, %s", strings.Join(addr, " "), res)
		}
	case pb.EnumRole_Candidate:
		logrus.Error("[switchor] elector (golang version) in master-slave mode should not use 'candidate'.")
	}

}

type mysqlOperator struct {
	sw *Switchor

	LocalAddr      string
	LocalUser      string
	LocalPassword  string
	RemoteAddr     string
	RemoteUser     string
	RemotePassword string

	ConnTimeout int
	SyncTimeout int

	ldb *sql.DB
	rdb *sql.DB
}

func newMySQLOperator(sw *Switchor, la, lu, lp, ra, ru, rp string, ct, st int) *mysqlOperator {
	return &mysqlOperator{
		sw: sw,

		LocalAddr:      la,
		LocalUser:      lu,
		LocalPassword:  lp,
		RemoteAddr:     ra,
		RemoteUser:     ru,
		RemotePassword: rp,

		ConnTimeout: ct,
		SyncTimeout: st,
	}
}

func (op *mysqlOperator) check_and_switch(electorRole pb.EnumRole) {
	match, err := op.check_mysql_role_match(electorRole)
	if err != nil {
		logrus.Errorf("[switchor] check_mysql_role_match() failed (err: %v), need to try again", err)
	} else {
		if !match {
			op.switch_mysql_role(electorRole)
		}
	}
}

func getReadOnly(db *sql.DB) (error, string) {
	var name, value string
	err := db.QueryRow("show global variables like 'read_only'").Scan(&name, &value)
	if err != nil {
		logrus.Warnf("[switchor] exec 'show global variables like \"read_only\"' failed, %v", err)
		return err, ""
	}

	return nil, value
}

func setReadOnly(db *sql.DB, to string) bool {
	_, err := db.Exec(fmt.Sprintf("set global read_only=%s", to))
	if err != nil {
		logrus.Warnf("[switchor] exec 'set global read_only=%s' failed, %v", to, err)
		return false
	} else {
		logrus.Infof("[switchor] exec 'set global read_only=%s' success", to)
		return true
	}
}

func (op *mysqlOperator) check_mysql_role_match(electorRole pb.EnumRole) (bool, error) {

	err, ro := getReadOnly(op.ldb)
	if err != nil {
		return false, err
	}

	var curAppRole appRole
	if ro == "ON" {
		curAppRole = appIsSlave
	} else {
		curAppRole = appIsMaster
	}

	logrus.Debugf("[switchor] <mysqlOperator> obtain mysql role ==> %v", appRole2str[curAppRole])

	return is_role_match("mysql", curAppRole, electorRole), nil
}

func masterReplStatus(db *sql.DB) (string, string) {
	var value [5]string
	err := db.QueryRow("show master status").Scan(&value)
	if err != nil {
		logrus.Warnf("[switchor] 'show master status' failed, %v", err)
		return "", ""
	} else {
		// mysql_master_repl_result_file_index = int32(0)
		// mysql_master_repl_result_pos_index  = int32(1)
		return value[0], value[1]
	}
}

func slaveReplStatus(db *sql.DB) (string, string) {
	var value [58]string
	err := db.QueryRow("show slave status").Scan(&value)
	if err != nil {
		logrus.Warnf("[switchor] 'show slave status' failed, %v", err)
		return "", ""
	} else {
		// mysql_slave_repl_result_file_index  = int32(5)
		// mysql_slave_repl_result_pos_index   = int32(21)
		return value[5], value[21]
	}
}

func (op *mysqlOperator) switch_mysql_role(electorRole pb.EnumRole) {

	switch electorRole {
	case pb.EnumRole_Leader:
		/*
		   steps to switch mysql to master:

		   1. check if read-only flag of remote mysql (old master) has been set (which deny all write ops to
		      prepare syncing):
		      if not, or shit happened to mysql, continue looping from 1.
		      if yes, go to 2.;
		   2. fetch {File:Position} of command `show master status` on the remote (old master) mysql;
		   3. fetch {Master_Log_File:Exec_Master_Log_Pos} of command `show slave status` on the local mysql;
		   4. wait until {File:Position} and {Master_Log_File:Exec_Master_Log_Pos} be the same, or timeout;
		   5. set read_only=0 no matter whether the sync progress finished or not.
		*/

		ticker := time.NewTicker(time.Duration(1) * time.Second)
		syncTimeout := time.After(time.Duration(op.SyncTimeout) * time.Second)

		for {
			select {
			case <-syncTimeout:
				logrus.Warnf("[switchor] sync not complete in %ds, but we still set local mysql 'read_only' to 'OFF'", op.SyncTimeout)
				setReadOnly(op.ldb, "OFF")
				return

			case <-ticker.C:
				var ro string
				err, ro := getReadOnly(op.rdb)
				if err != nil {
					continue
				}

				if ro == "ON" {
					// remote mysql 'read_only' is ON
					file, position := masterReplStatus(op.rdb)
					if file == "" || position == "" {
						continue
					}
					logrus.Infof("[switchor] remote mysql (master) --> File [%s] Position [%s]", file, position)

					masterLogFile, execMasterLogPos := slaveReplStatus(op.ldb)
					if masterLogFile == "" || execMasterLogPos == "" {
						continue
					}
					logrus.Infof("[switchor] local mysql (slave) --> Master_Log_File [%s] Exec_Master_Log_Pos [%s]", masterLogFile, execMasterLogPos)

					if file == masterLogFile && position == execMasterLogPos {
						logrus.Info("[switchor] mysql master-slave in SYNC state")
					} else {
						logrus.Info("[switchor] mysql master-slave in un-SYNC state")
					}

				} else {
					// NOTE: 原设计中，等待 remote mysql 切换 read_only 为 ON 的时间也被算在 SyncTimeout 时间之中了
					// 若 remote MySQL 迟迟未切换到 ON ，则真正留给等待同步的时间会很少，可能导致数据丢失
					logrus.Info("[switchor] remote mysql 'read_only' is still 'OFF', want for changing")
				}
			}
		}

	case pb.EnumRole_Follower:
		setReadOnly(op.ldb, "ON")
	case pb.EnumRole_Candidate:
		logrus.Error("[switchor] elector (golang version) in master-slave mode should not use 'candidate'.")
	}

}

// Switchor defines the switchor
type Switchor struct {
	radarHost  string // radar server tcp host
	rsTcpHost  string // role service tcp host
	rsUnixHost string // role service unix host

	checkPeriod     uint
	reconnectPeriod uint

	mode string

	operators map[string]Operator

	rsClientConn *grpc.ClientConn // connection to remote elector
	rsClient     pb.RoleServiceClient

	radarCli *radar.RadarClient

	disconnectedRadarCh   chan struct{}
	disconnectedElectorCh chan struct{}

	connectedRadarCh   chan struct{}
	connectedElectorCh chan struct{}

	roleNotifyCh chan pb.EnumRole

	stopCh chan struct{}
}

// NewSwitchor returns a switchor instance
func NewSwitchor() *Switchor {

	s := &Switchor{
		radarHost:  parser.SwitchorSetting.RadarHost,
		rsTcpHost:  parser.SwitchorSetting.ElectorRoleServiceTcpHost,
		rsUnixHost: parser.SwitchorSetting.ElectorRoleServiceUnixHost,

		checkPeriod:     parser.SwitchorSetting.CheckPeriod,
		reconnectPeriod: parser.SwitchorSetting.ReconnectPeriod,

		mode: parser.SwitchorSetting.Mode,
	}

	s.operators = map[string]Operator{}

	s.disconnectedRadarCh = make(chan struct{}, 1)
	s.disconnectedElectorCh = make(chan struct{}, 1)

	s.connectedRadarCh = make(chan struct{}, 1)
	s.connectedElectorCh = make(chan struct{}, 1)

	s.roleNotifyCh = make(chan pb.EnumRole)

	s.stopCh = make(chan struct{})

	return s
}

// Start the switchor
func (s *Switchor) Start() error {
	go s.electorLoop()
	go s.radarLoop()

	s.create_operators()
	s.start_operators()

	return nil
}

// Stop the switchor
func (s *Switchor) Stop() {
	close(s.stopCh)

	s.disconnectElector()
	s.disconnectRadar()
}

// FIXME: 可以进一步抽象
func (op *modbOperator) operatorLoop() {
	for {
		select {
		case <-op.sw.stopCh:
			return

		case curRole := <-op.sw.roleNotifyCh:
			logrus.Debugf("[switchor] <modbOperator> obtain elector curRole ==> %v", curRole)
			op.check_and_switch(curRole)
		}
	}
}

func (op *redisOperator) operatorLoop() {

	// NOTE: do connection job here avoiding to connect redis each loop
	c, err := redis.Dial("tcp", op.LocalAddr)
	if err != nil {
		logrus.Warnf("[switchor] connect Redis[%s] failed, %v", op.LocalAddr, err)
		return
	} else {
		logrus.Debugf("[switchor] connect Redis[%s] success", op.LocalAddr)
	}

	if op.LocalPassword != "" {
		if _, err := c.Do("AUTH", op.LocalPassword); err != nil {
			c.Close()
			logrus.Warnf("[switchor] redis AUTH failed, %v", err)
			return
		}
	}
	defer c.Close()

	op.c = c

	for {
		select {
		case <-op.sw.stopCh:
			return

		case curRole := <-op.sw.roleNotifyCh:
			logrus.Debugf("[switchor] <redisOperator> obtain elector curRole ==> %v", curRole)
			op.check_and_switch(curRole)
		}
	}
}

func (op *mysqlOperator) operatorLoop() {

	// NOTE: do connection job here avoiding to connect local and remote mysql each loop
	ldb, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/?timeout=%ds&charset=utf8&parseTime=True&loc=Local",
		op.LocalUser,
		op.LocalPassword,
		op.LocalAddr,
		op.ConnTimeout))
	if err != nil {
		logrus.Warnf("[switchor] connect local MySQL[%s] failed, %v", op.LocalAddr, err)
		return
	} else {
		logrus.Debugf("[switchor] connect local MySQL[%s] success", op.LocalAddr)
	}
	defer ldb.Close()

	op.ldb = ldb

	rdb, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/?timeout=%ds&charset=utf8&parseTime=True&loc=Local",
		op.RemoteUser,
		op.RemotePassword,
		op.RemoteAddr,
		op.ConnTimeout))
	if err != nil {
		logrus.Warnf("[switchor] connect remote MySQL[%s] failed, %v", op.RemoteAddr, err)
		return
	} else {
		logrus.Debugf("[switchor] connect remote MySQL[%s] success", op.RemoteAddr)
	}
	defer rdb.Close()

	op.rdb = rdb

	for {
		select {
		case <-op.sw.stopCh:
			return

		case curRole := <-op.sw.roleNotifyCh:
			logrus.Debugf("[switchor] <mysqlOperator> obtain elector curRole ==> %v", curRole)
			op.check_and_switch(curRole)
		}
	}
}

func (s *Switchor) start_operators() {

	if op, ok := s.operators["modb"]; ok {
		go op.(*modbOperator).operatorLoop()
	}
	if op, ok := s.operators["redis"]; ok {
		go op.(*redisOperator).operatorLoop()
	}
	if op, ok := s.operators["mysql"]; ok {
		go op.(*mysqlOperator).operatorLoop()
	}
}

func (s *Switchor) create_operators() {

	switch parser.SwitchorSetting.Mode {
	case "single-point":
		// NOTE: only for keeping modb state right when changing mode from
		// master-slave or cluster to single-point
		_, ok := parser.OperatorRegistry["modb"]
		if ok {
			op := newModbOperator(
				s,

				parser.ModbSetting.DomainMoid,
				parser.ModbSetting.MachineRoomMoid,
				parser.ModbSetting.GroupMoid,
				parser.ModbSetting.ServerMoid,
			)

			logrus.Infof("[switchor] ==> modb operator: %+v", op)
			s.operators["modb"] = op
		} else {
			logrus.Error("[switchor] the config of 'modb' operator MUST exist in single-point mode")
		}

	case "master-slave":
		var ok bool
		// modb
		_, ok = parser.OperatorRegistry["modb"]
		if ok {
			op := newModbOperator(
				s,

				parser.ModbSetting.DomainMoid,
				parser.ModbSetting.MachineRoomMoid,
				parser.ModbSetting.GroupMoid,
				parser.ModbSetting.ServerMoid,
			)

			logrus.Infof("[switchor] ==> modb operator: %+v", op)
			s.operators["modb"] = op
		} else {
			logrus.Error("[switchor] the config of 'modb' operator MUST exist in master-slave mode")
		}

		// redis
		_, ok = parser.OperatorRegistry["redis"]
		if ok {
			op := newRedisOperator(
				s,

				parser.RedisSetting.LocalAddr,
				parser.RedisSetting.LocalPassword,
				parser.RedisSetting.RemoteAddr,
				parser.RedisSetting.RemotePassword,
			)

			logrus.Infof("[switchor] ==> redis operator: %+v", op)
			s.operators["redis"] = op
		} else {
			logrus.Error("[switchor] the config of 'redis' operator MUST exist in master-slave mode")
		}

		// mysql
		_, ok = parser.OperatorRegistry["mysql"]
		if ok {
			op := newMySQLOperator(
				s,

				parser.MySQLSetting.LocalAddr,
				parser.MySQLSetting.LocalUser,
				parser.MySQLSetting.LocalPassword,
				parser.MySQLSetting.RemoteAddr,
				parser.MySQLSetting.RemoteUser,
				parser.MySQLSetting.RemotePassword,

				parser.MySQLSetting.ConnTimeout,
				parser.MySQLSetting.SyncTimeout,
			)

			logrus.Infof("[switchor] ==> mysql operator: %+v", op)
			s.operators["mysql"] = op
		} else {
			logrus.Error("[switchor] the config of 'mysql' operator MUST exist in master-slave mode")
		}
	case "cluster":
	default:
		logrus.Fatal("not match any of [single-point|master-slave|cluster].")
	}
}
