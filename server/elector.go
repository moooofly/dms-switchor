package server

import (
	"context"
	"time"

	pb "github.com/moooofly/dms-switchor/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func (p *Switchor) electorLoop() {
	logrus.Debug("====> enter electorLoop")
	defer logrus.Debug("====> leave electorLoop")

	go p.backgroundConnectElector()
	p.electorReconnectTrigger()

	ticker := time.NewTicker(time.Duration(p.checkPeriod) * time.Second)

	for {
		select {
		case <-p.stopCh:
			return

		case <-p.connectedElectorCh:
		}

		var curRole, prevRole pb.EnumRole

		for range ticker.C {

			logrus.Infof("[switchor]    --------")

			// step 1: 获取 elector role
			logrus.Infof("[switchor] --> send [Obtain] to elector")

			obRsp, err := p.rsClient.Obtain(context.Background(), &pb.ObtainReq{})
			if err != nil {
				logrus.Infof("[switchor] Obtain role failed: %v", err)
				p.electorReconnectTrigger()
				break
			}

			logrus.Infof("[switchor] <-- recv [Obtain] Resp, role => [%s]", obRsp.GetRole())

			// NOTE: 由于默认值的关系，preRole 初始值总是 Candidate
			prevRole, curRole = curRole, obRsp.GetRole()
			logrus.Infof("[switchor]     role status, last [%s] --> current [%s]", prevRole, curRole)

			// step 2: 发现自己所属的 elector 发生了 role 变更，则根据策略触发业务的 switch 动作

			logrus.Infof("[switchor]    --------")
		}
	}
}

func (p *Switchor) backgroundConnectElector() {
	for {
		select {
		case <-p.stopCh:
			return

		case <-p.disconnectedElectorCh:
		}

		for {

			logrus.Infof("[switchor] --> try to connect elector[%s]", p.rsTcpHost)

			conn, err := grpc.Dial(
				p.rsTcpHost,
				grpc.WithInsecure(),
				grpc.WithBlock(),
				grpc.WithTimeout(time.Second),
			)
			// NOTE: block + timeout
			if err != nil {
				logrus.Warnf("[switchor] connect elector failed, reason: %v", err)
			} else {
				logrus.Infof("[switchor] connect elector success")

				// NOTE: 顺序不能变
				p.rsClientConn = conn
				p.rsClient = pb.NewRoleServiceClient(conn)

				// FIXME: another way?
				p.connectedElectorCh <- struct{}{}

				break
			}

			time.Sleep(time.Second * time.Duration(p.reconnectPeriod))
		}
	}
}

func (p *Switchor) electorReconnectTrigger() {
	select {
	case p.disconnectedElectorCh <- struct{}{}:
		logrus.Debugf("[switchor] trigger connection to [elector]")
	default:
		logrus.Debugf("[switchor] connection process is ongoing")
	}
}

func (p *Switchor) disconnectElector() {
	if p.rsClientConn != nil {
		p.rsClientConn.Close()
	}
}
