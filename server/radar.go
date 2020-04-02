package server

import (
	"time"

	radar "github.com/moooofly/radar-go-client"
	"github.com/sirupsen/logrus"
)

func (s *Switchor) radarLoop() {
	logrus.Debug("====> enter radarLoop")
	defer logrus.Debug("====> leave radarLoop")

	go s.backgroundConnectRadar()
	s.radarReconnectTrigger()

loop:
	select {
	case <-s.stopCh:
		return
	case <-s.connectedRadarCh:
		goto loop
	}

	// FIXME: 对比 electorLoop ，这里似乎缺少主逻辑
	// radar 对 heartbeat 逻辑是否应该放在这里？
	// 由于 radar 的 heartbeat 处理编写到 client 库中了，因此
}

func (s *Switchor) backgroundConnectRadar() {
	for {
		select {
		case <-s.stopCh:
			return
		case <-s.disconnectedRadarCh:
		}

		for {

			logrus.Infof("[switchor] --> try to connect radar[%s]", s.radarHost)

			c := radar.NewRadarClient(s.radarHost, logrus.StandardLogger())

			// NOTE: block + timeout
			if err := c.Connect(); err != nil {
				logrus.Warnf("[switchor] ## connect radar failed, reason: %v", err)
			} else {
				logrus.Infof("[switchor] ## connect radar success")

				// NOTE: 顺序不能变
				s.radarCli = c
				s.connectedRadarCh <- struct{}{}

				break
			}

			time.Sleep(time.Second * time.Duration(s.reconnectPeriod))
		}
	}
}

func (s *Switchor) radarReconnectTrigger() {
	select {
	case s.disconnectedRadarCh <- struct{}{}:
		logrus.Debugf("[switchor] trigger connection to [radar]")
	default:
		logrus.Debugf("[switchor] connection process is ongoing")
	}
}

func (s *Switchor) disconnectRadar() {
	if s.radarCli != nil {
		s.radarCli.Disconnect()
	}
}
