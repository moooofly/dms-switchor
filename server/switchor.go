package server

import (
	"github.com/moooofly/dms-switchor/pkg/parser"
	"google.golang.org/grpc"

	radar "github.com/moooofly/radar-go-client"

	pb "github.com/moooofly/dms-switchor/proto"
)

// Switchor defines the switchor
type Switchor struct {
	radarHost  string // radar server tcp host
	rsTcpHost  string // role service tcp host
	rsUnixHost string // role service unix host

	checkPeriod     uint
	reconnectPeriod uint

	mode string

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

	p := &Switchor{
		radarHost: parser.SwitchorSetting.RadarHost,

		rsTcpHost:  parser.SwitchorSetting.ElectorRoleServiceTcpHost,
		rsUnixHost: parser.SwitchorSetting.ElectorRoleServiceUnixHost,

		checkPeriod:     parser.SwitchorSetting.CheckPeriod,
		reconnectPeriod: parser.SwitchorSetting.ReconnectPeriod,

		mode: parser.SwitchorSetting.Mode,
	}

	p.disconnectedRadarCh = make(chan struct{}, 1)
	p.disconnectedElectorCh = make(chan struct{}, 1)

	p.connectedRadarCh = make(chan struct{}, 1)
	p.connectedElectorCh = make(chan struct{}, 1)

	p.stopCh = make(chan struct{})

	return p
}

// Start the switchor
func (p *Switchor) Start() error {

	go p.radarLoop()
	go p.electorLoop()

	return nil
}

// Stop the switchor
func (p *Switchor) Stop() {
	close(p.stopCh)

	p.disconnectElector()
	p.disconnectRadar()
}
