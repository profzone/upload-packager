package consensus

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"longhorn/upload-packager/internal/constants/enum"
	"longhorn/upload-packager/pkg/storage"
	"sync"
)

type FSM struct {
	sync.RWMutex
	partitions            map[uint16]*storage.Partition
	raft                  *Raft
	snapshot              *snapshot
	partitionWriteIndex   []uint16
	partitionReadIndex    map[raft.ServerID][]uint16
	currentWritePartition int
	currentReadPartition  int
}

func (f *FSM) encodeToBytes(target interface{}, buffer *bytes.Buffer) error {
	partitionBuffer := bytes.NewBuffer(make([]byte, 0))
	encoder := gob.NewEncoder(partitionBuffer)
	err := encoder.Encode(target)
	if err != nil {
		return err
	}

	partitionLengthBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(partitionLengthBytes, uint64(len(partitionBuffer.Bytes())))
	_, err = buffer.Write(partitionLengthBytes)
	if err != nil {
		return err
	}
	_, err = buffer.Write(partitionBuffer.Bytes())
	return err
}

func (f *FSM) decodeFromBytes(target interface{}, buffer *bytes.Reader) error {
	lengthBytes := make([]byte, 8)
	_, err := buffer.Read(lengthBytes)
	if err != nil {
		return err
	}
	partitionLength := binary.BigEndian.Uint64(lengthBytes)
	partitionBytes := make([]byte, partitionLength)
	_, err = buffer.Read(partitionBytes)
	if err != nil {
		return err
	}

	partitionReader := bytes.NewReader(partitionBytes)
	decoder := gob.NewDecoder(partitionReader)
	err = decoder.Decode(target)
	return err
}

func (f *FSM) GobDecode(data []byte) error {
	buf := bytes.NewReader(data)

	// decode partitions
	err := f.decodeFromBytes(&f.partitions, buf)
	if err != nil {
		return err
	}

	// decode partitionWriteIndex
	err = f.decodeFromBytes(&f.partitionWriteIndex, buf)
	if err != nil {
		return err
	}

	// decode partitionReadIndex
	err = f.decodeFromBytes(&f.partitionReadIndex, buf)
	if err != nil {
		return err
	}

	// decode currentWritePartition
	lengthBytes := make([]byte, 8)
	_, err = buf.Read(lengthBytes)
	if err != nil {
		return err
	}
	f.currentWritePartition = int(binary.BigEndian.Uint64(lengthBytes))

	// decode currentWritePartition
	_, err = buf.Read(lengthBytes)
	if err != nil {
		return err
	}
	f.currentReadPartition = int(binary.BigEndian.Uint64(lengthBytes))

	return nil
}

func (f *FSM) GobEncode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))

	// encode partitions
	err := f.encodeToBytes(f.partitions, buf)
	if err != nil {
		return nil, err
	}

	// encode partitionWriteIndex
	err = f.encodeToBytes(f.partitionWriteIndex, buf)
	if err != nil {
		return nil, err
	}

	// encode partitionReadIndex
	err = f.encodeToBytes(f.partitionReadIndex, buf)
	if err != nil {
		return nil, err
	}

	// encode currentWritePartition
	currentWritePartitionBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(currentWritePartitionBytes, uint64(f.currentWritePartition))
	_, err = buf.Write(currentWritePartitionBytes)
	if err != nil {
		return nil, err
	}

	// encode currentReadPartition
	currentReadPartitionBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(currentReadPartitionBytes, uint64(f.currentReadPartition))
	_, err = buf.Write(currentReadPartitionBytes)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func NewFSM(r *Raft) *FSM {
	fsm := &FSM{
		raft:                r,
		partitions:          make(map[uint16]*storage.Partition),
		partitionWriteIndex: make([]uint16, 0),
		partitionReadIndex:  make(map[raft.ServerID][]uint16),
	}
	snapshot := newSnapshot(fsm)
	fsm.snapshot = snapshot

	return fsm
}

func (f *FSM) InitPartitions(count uint16) bool {
	// if length of the partitions is bigger than 0, it represents that restore proceeded
	if len(f.partitions) > 0 {
		return false
	}
	var index uint16 = 0
	serverID := f.raft.GetServerID()
	f.partitionReadIndex[serverID] = make([]uint16, 0)
	for index < count {
		p := storage.NewPartition(index)
		f.partitions[index] = p
		f.partitionWriteIndex = append(f.partitionWriteIndex, index)
		f.partitionReadIndex[serverID] = append(f.partitionReadIndex[serverID], index)
		index++
	}

	return true
}

// when call this method, you need synced call
func (f *FSM) getNextReadPartition() *storage.Partition {
	var p *storage.Partition
	var partitionCount = len(f.partitions)
	var partitionReadCount = len(f.partitionReadIndex[f.raft.GetServerID()])
	for i := 0; i < partitionCount; i++ {
		f.currentReadPartition++
		if f.currentReadPartition >= partitionReadCount {
			f.currentReadPartition = 0
		}

		p = f.partitions[f.partitionReadIndex[f.raft.GetServerID()][f.currentReadPartition]]
		if p.Length() > 0 {
			return p
		}
	}
	return nil
}

// when call this method, you need synced call
func (f *FSM) getNextWritePartition() *storage.Partition {
	if f.currentWritePartition >= len(f.partitionWriteIndex) {
		f.currentWritePartition = 0
	}

	p := f.partitions[f.partitionWriteIndex[f.currentWritePartition]]
	f.currentWritePartition++
	return p
}

func (f *FSM) Pop() ([]byte, error) {
	f.RLock()
	defer f.RUnlock()

	p := f.getNextReadPartition()
	if p == nil {
		return nil, fmt.Errorf("no available partition")
	}
	return p.Pop()
}

func (f *FSM) Push(data []byte) {
	f.Lock()
	defer f.Unlock()

	p := f.getNextWritePartition()
	p.Push(data)
}

func (f *FSM) Marshal() (result []byte, err error) {
	buf := bytes.NewBuffer(result)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(f)
	return buf.Bytes(), err
}

func (f *FSM) Unmarshal(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(f)
	return
}

func (f *FSM) Apply(logEntryEvent *raft.Log) interface{} {
	entry := &LogEntryEvent{}
	err := entry.Unmarshal(logEntryEvent.Data)
	if err != nil {
		return err
	}

	switch entry.Type {
	case enum.EVENT_TYPE__MESSAAGE:
		f.Push(entry.Payload)
	case enum.EVENT_TYPE__PARTITION:
		count := binary.BigEndian.Uint16(entry.Payload)
		f.InitPartitions(count)
	}
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return f.snapshot, nil
}

func (f *FSM) Restore(reader io.ReadCloser) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	err = f.Unmarshal(data)
	logrus.Debug("restore from snapshot")
	return err
}
