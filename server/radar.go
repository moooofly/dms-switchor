package server

import (
	"time"

	radar "github.com/moooofly/radar-go-client"
	"github.com/sirupsen/logrus"
)

func (p *Switchor) radarLoop() {
	logrus.Debug("====> enter radarLoop")
	defer logrus.Debug("====> leave radarLoop")

	go p.backgroundConnectRadar()
	p.radarReconnectTrigger()

loop:
	select {
	case <-p.stopCh:
		return

	case <-p.connectedRadarCh:

		goto loop
	}
}

func (p *Switchor) backgroundConnectRadar() {
	for {
		select {
		case <-p.stopCh:
			return

		case <-p.disconnectedRadarCh:
		}

		for {

			logrus.Infof("[switchor] --> try to connect radar[%s]", p.radarHost)

			c := radar.NewRadarClient(p.radarHost, logrus.StandardLogger())

			// NOTE: block + timeout
			if err := c.Connect(); err != nil {
				logrus.Warnf("[switchor] connect radar failed, reason: %v", err)
			} else {
				logrus.Infof("[switchor] connect radar success")

				// NOTE: 顺序不能变
				p.radarCli = c
				p.connectedRadarCh <- struct{}{}

				break
			}

			time.Sleep(time.Second * time.Duration(p.reconnectPeriod))
		}
	}
}

func (p *Switchor) radarReconnectTrigger() {
	select {
	case p.disconnectedRadarCh <- struct{}{}:
		logrus.Debugf("[switchor] trigger connection to [radar]")
	default:
		logrus.Debugf("[switchor] connection process is ongoing")
	}
}

func (p *Switchor) disconnectRadar() {
	if p.radarCli != nil {
		p.radarCli.Close()
	}
}
