package consensus

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/profzone/eden-framework/pkg/courier/client"
	"longhorn/upload-packager/internal/clients/client_peer"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Raft struct {
	serverID    raft.ServerID
	node        *raft.Raft
	fsm         *FSM
	stateNotify chan bool
	isLeader    bool

	ListenAddr        string
	DataDir           string
	DataPrefix        string
	BootstrapAsLeader bool
	JoinAddr          string
}

func (r *Raft) Init() error {
	r.stateNotify = make(chan bool, 1)

	go func() {
		for {
			select {
			case r.isLeader = <-r.stateNotify:
				color.Red("Leadership is changed: %v", r.isLeader)
			}
		}
	}()

	r.serverID = raft.ServerID(r.ListenAddr)
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = r.serverID
	raftConfig.SnapshotInterval = 20 * time.Second
	raftConfig.SnapshotThreshold = 2
	raftConfig.NotifyCh = r.stateNotify

	r.fsm = NewFSM(r)

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(r.DataDir, fmt.Sprintf("%s%s", r.DataPrefix, "raft-log.db")))
	if err != nil {
		return err
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(r.DataDir, fmt.Sprintf("%s%s", r.DataPrefix, "raft-stable.db")))
	if err != nil {
		return err
	}
	snapshotStore, err := raft.NewFileSnapshotStore(r.DataDir, 1, os.Stderr)
	if err != nil {
		return err
	}
	transport, err := r.newRaftTransport()
	if err != nil {
		return err
	}

	r.node, err = raft.NewRaft(raftConfig, r.fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return err
	}

	if r.BootstrapAsLeader {
		config := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		r.node.BootstrapCluster(config)
		r.fsm.InitPartitions(2)
	} else {
		address, err := net.ResolveTCPAddr("tcp", r.JoinAddr)
		if err != nil {
			return err
		}
		peerClient := &client_peer.ClientPeer{
			Client: client.Client{
				Host: address.IP.String(),
				Port: int16(address.Port),
			},
		}
		peerClient.MarshalDefaults(peerClient)

		createRequest := client_peer.CreatePeerRequest{
			Body: client_peer.CreatePeerBody{
				PeerID:   string(raftConfig.LocalID),
				PeerAddr: string(transport.LocalAddr()),
			},
		}
		_, err = peerClient.CreatePeer(createRequest)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Raft) newRaftTransport() (*raft.NetworkTransport, error) {
	address, err := net.ResolveTCPAddr("tcp", r.ListenAddr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(address.String(), address, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}
	return transport, nil
}

func (r *Raft) GetServerID() raft.ServerID {
	return r.serverID
}

func (r *Raft) Apply(command []byte, timeout time.Duration) raft.ApplyFuture {
	return r.node.Apply(command, timeout)
}

func (r *Raft) AddVoter(id string, address string, prevIndex uint64, timeout time.Duration) raft.IndexFuture {
	return r.node.AddVoter(raft.ServerID(id), raft.ServerAddress(address), prevIndex, timeout)
}

func (r *Raft) IsLeader() bool {
	return r.isLeader
}

func (r *Raft) Pop() ([]byte, error) {
	return r.fsm.Pop()
}

func (r *Raft) Push(data []byte) error {
	if !r.isLeader {
		return fmt.Errorf("only leader can set data")
	}

	applyFuture := r.Apply(data, 10*time.Second)
	if err := applyFuture.Error(); err != nil {
		return err
	}

	return nil
}
