package server

import (
	"context"
	"time"

	"github.com/moooofly/dms-switchor/pkg/parser"
	pb "github.com/moooofly/dms-switchor/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func (s *Switchor) electorLoop() {
	logrus.Debug("====> enter electorLoop")
	defer logrus.Debug("====> leave electorLoop")

	go s.backgroundConnectElector()
	s.electorReconnectTrigger()

	ticker := time.NewTicker(time.Duration(s.checkPeriod) * time.Second)

	for {
		select {
		case <-s.stopCh:
			return

		case <-s.connectedElectorCh:
		}

		var curRole, prevRole pb.EnumRole
		var counter uint = 0

		for range ticker.C {

			logrus.Infof("[switchor]    --------")

			// step 1: 获取 elector role
			logrus.Infof("[switchor] --> send [Obtain] to elector")

			obRsp, err := s.rsClient.Obtain(context.Background(), &pb.ObtainReq{})
			if err != nil {
				logrus.Infof("[switchor] Obtain role failed: %v", err)
				s.electorReconnectTrigger()
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

					// NOTE: notify curRole if could, or discard
					select {
					case s.roleNotifyCh <- curRole:
					default:
					}

					logrus.Infof("[switchor] ------> counter => %d (> threshold (%d))", counter, parser.SwitchorSetting.SwitchThreshold)
				}
			}

			logrus.Infof("[switchor]    --------")
		}
	}
}

func (s *Switchor) backgroundConnectElector() {
	for {
		select {
		case <-s.stopCh:
			return

		case <-s.disconnectedElectorCh:
		}

		for {

			logrus.Infof("[switchor] --> try to connect elector[%s]", s.rsTcpHost)

			conn, err := grpc.Dial(
				s.rsTcpHost,
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
				s.rsClientConn = conn
				s.rsClient = pb.NewRoleServiceClient(conn)

				// FIXME: another way?
				s.connectedElectorCh <- struct{}{}

				break
			}

			time.Sleep(time.Second * time.Duration(s.reconnectPeriod))
		}
	}
}

func (s *Switchor) electorReconnectTrigger() {
	select {
	case s.disconnectedElectorCh <- struct{}{}:
		logrus.Debugf("[switchor] trigger connection to [elector]")
	default:
		logrus.Debugf("[switchor] connection process is ongoing")
	}
}

func (s *Switchor) disconnectElector() {
	if s.rsClientConn != nil {
		s.rsClientConn.Close()
	}
}
