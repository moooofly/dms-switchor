package server

import (
	"context"
	"time"

	"github.com/moooofly/dms-switchor/pkg/parser"
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
		var counter uint = 0

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

			// step 2: 若自己所属的 elector 发生了 role 变更，则根据策略触发业务的 switch 动作
			// NOTE: 由于默认值的关系，preRole 初始值总是 Candidate
			// FIXME: 考虑设置初始值为 Follower
			prevRole, curRole = curRole, obRsp.GetRole()
			if prevRole != curRole {
				logrus.Infof("[switchor]     role change, last [%s] --> current [%s]", prevRole, curRole)
				counter = 1
			} else {
				counter += 1
				if counter >= parser.SwitchorSetting.SwitchThreshold {
					// TODO:通过 channel 发送信号给 operators
					logrus.Infof("[switchor] ------> counter => %d (> threshold (%d))", counter, parser.SwitchorSetting.SwitchThreshold)
				}
			}

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
