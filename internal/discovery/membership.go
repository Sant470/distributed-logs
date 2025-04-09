/* The specific job we want our discovery layer to solve is to
tell us when a server joined or left the cluster and what's its IDs and Addresses.

To implement a service discovery with the Serf we need to:
1. Create a serf node on each server
2. Configure each serf node with an address to listen on and accept connections from other Serf Nodes
3. Configure each Serf node with addresses of other serf nodes and join the cluster.
4. Handle Serf's cluster discovery events, such as when a node joins or fails in the cluster.

*/

package discovery

import (
	"net"

	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"
)

type Config struct {
	NodeName       string // node's unique identifier, if not provided Serf uses the hostname
	BindAddr       string
	Tags           map[string]string
	StartJoinAddrs []string
}

type Handler interface {
	Join(name, addr string) error
	Leave(name string) error
}

/* wrapper over serf to provide discovery and cluster membership*/
type Membership struct {
	Config
	handler Handler
	serf    *serf.Serf
	events  chan serf.Event
	logger  *zap.Logger
}

func New(h Handler, c Config) (*Membership, error) {
	m := &Membership{
		Config:  c,
		handler: h,
		logger:  zap.L().Named("membership"),
	}
	if err := m.setupSerf(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Membership) setupSerf() error {
	addr, err := net.ResolveTCPAddr("tcp", m.BindAddr)
	if err != nil {
		return err
	}
	serfConfig := serf.DefaultConfig()
	serfConfig.Init()
	// Serf listen on these addresses and port for gossiping
	serfConfig.MemberlistConfig.BindAddr = addr.IP.String()
	serfConfig.MemberlistConfig.BindPort = addr.Port
	m.events = make(chan serf.Event)
	// communicate events related to node joining and leaving the cluster
	serfConfig.EventCh = m.events
	// tags is to share each node's user-configured RPC address
	serfConfig.Tags = m.Tags
	serfConfig.NodeName = m.NodeName
	m.serf, err = serf.Create(serfConfig)
	if err != nil {
		return err
	}
	go m.eventHandler()
	if m.StartJoinAddrs != nil {
		_, err = m.serf.Join(m.StartJoinAddrs, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Membership) eventHandler() {
	for e := range m.events {
		switch e.EventType() {
		case serf.EventMemberJoin:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					continue
				}
				m.handleJoin(member)
			}
		case serf.EventMemberFailed, serf.EventMemberLeave:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					continue
				}
				m.handleLeave(member)
			}
		}
	}
}

func (m *Membership) isLocal(member serf.Member) bool {
	return m.serf.LocalMember().Name == member.Name
}

func (m *Membership) handleJoin(member serf.Member) {
	if err := m.handler.Join(member.Name, member.Tags["rpc_addr"]); err != nil {
		m.logError(err, "failed to join", member)
	}
}

func (m *Membership) handleLeave(member serf.Member) {
	if err := m.handler.Leave(member.Name); err != nil {
		m.logError(err, "failed to leave", member)
	}
}

func (m *Membership) logError(err error, message string, member serf.Member) {
	m.logger.Error(message, zap.Error(err), zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
}

// return a point in time snapshot of the cluster's serf members
func (m *Membership) Members() []serf.Member {
	return m.serf.Members()
}

func (m *Membership) Leave() error {
	return m.serf.Leave()
}
