package server

import (
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
// NOTE: 该函数应该命名为 is_role_match() 之类的
func is_role_match(appName string, ar appRole, er pb.EnumRole) bool {
	if appRole2electorRole[ar] == er {
		logrus.Info("[switchor] [%s] appRole(%s) and electorRole(%s) are Matched.", appName, appRole2str[ar], er)
		return true
	} else {
		logrus.Info("[switchor] [%s] appRole(%s) and electorRole(%s) are Mismatched.", appName, appRole2str[ar], er)
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
	radarHost string
	radarCli  *radar.RadarClient

	DomainMoid      string
	MachineRoomMoid string // equal to resourceMoid
	GroupMoid       string
	ServerMoid      string // moid
}

func newModbOperator(rhost, dm, rm, gm, sm string) *modbOperator {
	return &modbOperator{
		radarHost:       rhost,
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
	rspGroup, err := op.radarCli.GetAppControl([]radar.ReqArgs{a})
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
	// NOTE: 这应该打印 modb 的 role ，而不是直接 elector role
	logrus.Info("[switchor] switch role of modb to '%v'", electorRole)

	var oper string
	if electorRole == pb.EnumRole_Leader {
		oper = "1" // enabled
	} else {
		oper = "0" // disabled
	}

	// FIXME: 需要对 radarCli 是否可用做判定，并触发重连
	ok, err := op.radarCli.SetAppControl(oper, op.DomainMoid, op.MachineRoomMoid, op.GroupMoid, op.ServerMoid, "modb")
	if err != nil {
		logrus.Errorf("[switchor] SetAppControl failed: %v", err)
		return
	}

	if ok {
		logrus.Info("[switchor] switch role of modb success")
	} else {
		logrus.Warn("[switchor] switch role of modb failed")
	}
}

type redisOperator struct {
	LocalAddr      string
	LocalPassword  string
	RemoteAddr     string
	RemotePassword string
}

func newRedisOperator(la, lp, ra, rp string) *redisOperator {
	return &redisOperator{
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

func (op *redisOperator) check_redis_role_match(electorRole pb.EnumRole) (bool, error) {
	return false, nil
}
func (op *redisOperator) switch_redis_role(electorRole pb.EnumRole) {
}

type mysqlOperator struct {
	LocalAddr      string
	LocalUser      string
	LocalPassword  string
	RemoteAddr     string
	RemoteUser     string
	RemotePassword string

	ConnTimeout int
	SyncTimeout int
}

func newMySQLOperator(la, lu, lp, ra, ru, rp string, ct, st int) *mysqlOperator {
	return &mysqlOperator{
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

func (op *mysqlOperator) check_mysql_role_match(electorRole pb.EnumRole) (bool, error) {
	return false, nil
}
func (op *mysqlOperator) switch_mysql_role(electorRole pb.EnumRole) {
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

	s.stopCh = make(chan struct{})

	return s
}

// Start the switchor
func (s *Switchor) Start() error {
	go s.electorLoop()
	go s.radarLoop()

	// FIXME: 放在这里 or 放在 NewSwitchor 中？
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

func (op *modbOperator) operatorLoop() {
	// 获取当前 electorRole
	var electorRole pb.EnumRole
	op.check_and_switch(electorRole)
}

func (op *redisOperator) operatorLoop() {
	// 获取当前 electorRole
	var electorRole pb.EnumRole
	op.check_and_switch(electorRole)
}

func (op *mysqlOperator) operatorLoop() {
	// 获取当前 electorRole
	var electorRole pb.EnumRole
	op.check_and_switch(electorRole)
}

func (s *Switchor) start_operators() {
	// 针对每个 operator 启动一个 goroutine 运行对应的 check_and_switch
	// 之后通过 channel 触发调用

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
				parser.SwitchorSetting.RadarHost,
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
				parser.SwitchorSetting.RadarHost,

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
