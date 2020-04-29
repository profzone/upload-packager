package consensus

import (
	"github.com/hashicorp/raft"
)

type snapshot struct {
	fsm *FSM
}

func newSnapshot(fsm *FSM) *snapshot {
	return &snapshot{
		fsm: fsm,
	}
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	data, err := s.fsm.Marshal()
	if err != nil {
		sink.Cancel()
		return err
	}
	if _, err := sink.Write(data); err != nil {
		sink.Cancel()
		return err
	}
	if err := sink.Close(); err != nil {
		sink.Cancel()
		return err
	}
	return nil
}

func (s *snapshot) Release() {

}
