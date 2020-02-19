package parser

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

var (
	cfg *ini.File

	SwitchorSetting = &switchor{}
	MySQLSetting    = &mysql{}
	RedisSetting    = &redis{}
	ModbSetting     = &modb{}
)

// [switchor] section in .ini
type switchor struct {
	RadarHost string `ini:"radar-server-host"`

	ElectorRoleServiceTcpHost  string `ini:"elector-role-service-host"`
	ElectorRoleServiceUnixHost string `ini:"elector-role-service-path"`

	LogPath  string `ini:"log-path"`
	LogLevel string `ini:"log-level"`

	CheckPeriod     uint `ini:"elector-role-check-period"`
	ReconnectPeriod uint `ini:"radar-reconnect-period"`

	Mode string `ini:"mode"`
}

// [redis] section in .ini
type redis struct {
	LocalAddr      string `ini:"redis-local-addr"`
	LocalPassword  string `ini:"redis-local-password"`
	RemoteAddr     string `ini:"redis-remote-addr"`
	RemotePassword string `ini:"redis-remote-password"`
}

// [mysql] section in .ini
type mysql struct {
	LocalAddr      string `ini:"mysql-local-addr"`
	LocalUser      string `ini:"mysql-local-user"`
	LocalPassword  string `ini:"mysql-local-password"`
	RemoteAddr     string `ini:"mysql-remote-addr"`
	RemoteUser     string `ini:"mysql-remote-user"`
	RemotePassword string `ini:"mysql-remote-password"`

	ConnTimeout int `ini:"mysql-conn-timeout"`
	SyncTimeout int `ini:"mysql-sync-timeout"`
}

// [modb] section in .ini
type modb struct {
	DomainMoid      string `ini:"domain-moid"`
	MachineRoomMoid string `ini:"machine-room-moid"`
	GroupMoid       string `ini:"group-moid"`
	Moid            string `ini:"moid"`
}

func Load() {
	// TODO: 路径问题
	var err error
	cfg, err = ini.Load("conf/switchor.ini")
	if err != nil {
		logrus.Fatalf("Fail to parse 'conf/switchor.ini': %v", err)
	}

	mapTo("switchor", SwitchorSetting)
	mapTo("redis", RedisSetting)
	mapTo("mysql", MySQLSetting)
	mapTo("modb", ModbSetting)

	switch SwitchorSetting.Mode {
	case "single-point":
		logrus.Infof("mode => [%s]", SwitchorSetting.Mode)
	case "master-slave":
		logrus.Infof("mode => [%s]", SwitchorSetting.Mode)
	case "cluster":
		logrus.Infof("mode => [%s]", SwitchorSetting.Mode)
	default:
		logrus.Fatal("not match any of [single-point|master-slave|cluster].")
	}
}

func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		logrus.Fatalf("mapto err: %v", err)
	}
}
