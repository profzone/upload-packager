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

func (F *FSM) GobDecode(data []byte) error {
	buf := bytes.NewReader(data)

	// decode partitions
	lengthBytes := make([]byte, 8)
	_, err := buf.Read(lengthBytes)
	if err != nil {
		return err
	}
	partitionLength := binary.BigEndian.Uint64(lengthBytes)
	partitionBytes := make([]byte, partitionLength)
	_, err = buf.Read(partitionBytes)
	if err != nil {
		return err
	}

	partitionReader := bytes.NewReader(partitionBytes)
	decoder := gob.NewDecoder(partitionReader)
	err = decoder.Decode(&F.partitions)
	if err != nil {
		return err
	}

	// decode partitionWriteIndex
	_, err = buf.Read(lengthBytes)
	if err != nil {
		return err
	}
	partitionWriteIndexLength := binary.BigEndian.Uint64(lengthBytes)
	partitionWriteIndexBytes := make([]byte, partitionWriteIndexLength)
	_, err = buf.Read(partitionWriteIndexBytes)
	if err != nil {
		return err
	}

	partitionWriteIndexBytesReader := bytes.NewReader(partitionWriteIndexBytes)
	decoder = gob.NewDecoder(partitionWriteIndexBytesReader)
	err = decoder.Decode(&F.partitionWriteIndex)
	if err != nil {
		return err
	}

	// decode partitionReadIndex
	_, err = buf.Read(lengthBytes)
	if err != nil {
		return err
	}
	partitionReadIndexLength := binary.BigEndian.Uint64(lengthBytes)
	partitionReadIndexBytes := make([]byte, partitionReadIndexLength)
	_, err = buf.Read(partitionReadIndexBytes)
	if err != nil {
		return err
	}

	partitionReadIndexBytesReader := bytes.NewReader(partitionReadIndexBytes)
	decoder = gob.NewDecoder(partitionReadIndexBytesReader)
	err = decoder.Decode(&F.partitionReadIndex)
	if err != nil {
		return err
	}

	// decode currentWritePartition
	_, err = buf.Read(lengthBytes)
	if err != nil {
		return err
	}
	F.currentWritePartition = int(binary.BigEndian.Uint64(lengthBytes))

	// decode currentWritePartition
	_, err = buf.Read(lengthBytes)
	if err != nil {
		return err
	}
	F.currentReadPartition = int(binary.BigEndian.Uint64(lengthBytes))

	return nil
}

func (F *FSM) GobEncode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	partitionBuffer := bytes.NewBuffer(make([]byte, 0))
	encoder := gob.NewEncoder(partitionBuffer)

	// encode partitions
	err := encoder.Encode(F.partitions)
	if err != nil {
		return nil, err
	}

	partitionLengthBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(partitionLengthBytes, uint64(len(partitionBuffer.Bytes())))
	_, err = buf.Write(partitionLengthBytes)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(partitionBuffer.Bytes())
	if err != nil {
		return nil, err
	}

	// encode partitionWriteIndex
	partitionBuffer = bytes.NewBuffer(make([]byte, 0))
	encoder = gob.NewEncoder(partitionBuffer)
	err = encoder.Encode(F.partitionWriteIndex)
	if err != nil {
		return nil, err
	}

	partitionWriteIndexLengthBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(partitionWriteIndexLengthBytes, uint64(len(partitionBuffer.Bytes())))
	_, err = buf.Write(partitionWriteIndexLengthBytes)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(partitionBuffer.Bytes())
	if err != nil {
		return nil, err
	}

	// encode partitionReadIndex
	partitionBuffer = bytes.NewBuffer(make([]byte, 0))
	encoder = gob.NewEncoder(partitionBuffer)
	err = encoder.Encode(F.partitionReadIndex)
	if err != nil {
		return nil, err
	}

	partitionReadIndexLengthBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(partitionReadIndexLengthBytes, uint64(len(partitionBuffer.Bytes())))
	_, err = buf.Write(partitionReadIndexLengthBytes)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(partitionBuffer.Bytes())
	if err != nil {
		return nil, err
	}

	// encode currentWritePartition
	currentWritePartitionBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(currentWritePartitionBytes, uint64(F.currentWritePartition))
	_, err = buf.Write(currentWritePartitionBytes)
	if err != nil {
		return nil, err
	}

	// encode currentReadPartition
	currentReadPartitionBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(currentReadPartitionBytes, uint64(F.currentReadPartition))
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

func (F *FSM) InitPartitions(count uint16) bool {
	// if length of the partitions is bigger than 0, it represents that restore proceeded
	if len(F.partitions) > 0 {
		return false
	}
	var index uint16 = 0
	serverID := F.raft.GetServerID()
	F.partitionReadIndex[serverID] = make([]uint16, 0)
	for index < count {
		p := storage.NewPartition(index)
		F.partitions[index] = p
		F.partitionWriteIndex = append(F.partitionWriteIndex, index)
		F.partitionReadIndex[serverID] = append(F.partitionReadIndex[serverID], index)
		index++
	}

	return true
}

// when call this method, you need synced call
func (F *FSM) getNextReadPartition() *storage.Partition {
	var p *storage.Partition
	var partitionCount = len(F.partitions)
	var partitionReadCount = len(F.partitionReadIndex[F.raft.GetServerID()])
	for i := 0; i < partitionCount; i++ {
		if F.currentReadPartition >= partitionReadCount {
			F.currentReadPartition = 0
		}

		p = F.partitions[F.partitionReadIndex[F.raft.GetServerID()][F.currentReadPartition]]
		if p.Length() > 0 {
			break
		}
		F.currentReadPartition++
	}
	return p
}

// when call this method, you need synced call
func (F *FSM) getNextWritePartition() *storage.Partition {
	if F.currentWritePartition >= len(F.partitionWriteIndex) {
		F.currentWritePartition = 0
	}

	p := F.partitions[F.partitionWriteIndex[F.currentWritePartition]]
	F.currentWritePartition++
	return p
}

func (F *FSM) Pop() ([]byte, error) {
	F.RLock()
	defer F.RUnlock()

	p := F.getNextReadPartition()
	if p == nil {
		return nil, fmt.Errorf("no available partition")
	}
	return p.Pop()
}

func (F *FSM) Push(data []byte) {
	F.Lock()
	defer F.Unlock()

	p := F.getNextWritePartition()
	p.Push(data)
}

func (F *FSM) Marshal() (result []byte, err error) {
	buf := bytes.NewBuffer(result)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(F)
	return buf.Bytes(), err
}

func (F *FSM) Unmarshal(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(F)
	return
}

func (F *FSM) Apply(logEntryEvent *raft.Log) interface{} {
	entry := &LogEntryEvent{}
	err := entry.Unmarshal(logEntryEvent.Data)
	if err != nil {
		return err
	}

	switch entry.Type {
	case enum.EVENT_TYPE__MESSAAGE:
		F.Push(entry.Payload)
	case enum.EVENT_TYPE__PARTITION:
		count := binary.BigEndian.Uint16(entry.Payload)
		F.InitPartitions(count)
	}
	return nil
}

func (F *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return F.snapshot, nil
}

func (F *FSM) Restore(reader io.ReadCloser) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	err = F.Unmarshal(data)
	logrus.Debug("restore from snapshot")
	return err
}
