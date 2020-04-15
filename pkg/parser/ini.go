package parser

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

var (
	cfg *ini.File

	SwitchorSetting = &switchor{}

	// operator setting
	MySQLSetting = &mysql{}
	RedisSetting = &redis{}
	ModbSetting  = &modb{}

	OperatorRegistry = make(map[string]bool)
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

	SwitchThreshold uint `ini:"switch-threshold"`
}

// [redis] section in .ini
type redis struct {
	LocalAddr      string `ini:"redis-local-addr"`
	LocalPassword  string `ini:"redis-local-passwd"`
	RemoteAddr     string `ini:"redis-remote-addr"`
	RemotePassword string `ini:"redis-remote-passwd"`
}

// [mysql] section in .ini
type mysql struct {
	LocalAddr      string `ini:"mysql-local-addr"`
	LocalUser      string `ini:"mysql-local-user"`
	LocalPassword  string `ini:"mysql-local-passwd"`
	RemoteAddr     string `ini:"mysql-remote-addr"`
	RemoteUser     string `ini:"mysql-remote-user"`
	RemotePassword string `ini:"mysql-remote-passwd"`

	ConnTimeout int `ini:"mysql-conn-timeout"`
	SyncTimeout int `ini:"mysql-sync-timeout"`
}

// [modb] section in .ini
type modb struct {
	DomainMoid      string `ini:"domain-moid"`
	MachineRoomMoid string `ini:"machine-room-moid"` // alias to resource-moid
	GroupMoid       string `ini:"group-moid"`
	ServerMoid      string `ini:"server-moid"`
}

func Load() {
	// TODO: 路径问题
	var err error
	cfg, err = ini.Load("conf/switchor.ini")
	if err != nil {
		logrus.Fatalf("Fail to parse 'conf/switchor.ini': %v", err)
	}

	err = cfg.Section("switchor").MapTo(SwitchorSetting)
	if err != nil {
		logrus.Fatalf("MapTo(SwitchorSetting) failed: %v", err)
	}

	mapTo("redis", RedisSetting)
	mapTo("mysql", MySQLSetting)
	mapTo("modb", ModbSetting)

	logrus.Infof("operator registered => %v", OperatorRegistry)

	switch SwitchorSetting.Mode {
	case "single-point", "master-slave", "cluster":
		logrus.Infof("mode setting => [%s]", SwitchorSetting.Mode)
	default:
		logrus.Fatalf("mode '%v' dose not match any of [single-point|master-slave|cluster].", SwitchorSetting.Mode)
	}
}

func mapTo(section string, v interface{}) {
	sect, err := cfg.GetSection(section)
	if err != nil {
		logrus.Warnf("GetSection(%s) failed: %v", section, err)
	} else {
		if err = sect.MapTo(v); err != nil {
			logrus.Fatalf("MapTo() failed: %v", err)
		} else {
			OperatorRegistry[section] = true
		}
	}

}
