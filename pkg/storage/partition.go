package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
)

// minQueueLen is smallest capacity that queue may have.
// Must be power of 2 for bitwise modulus: x % n == x & (n - 1).
const minQueueLen = 16

type Partition struct {
	Index             uint16
	events            [][]byte
	head, tail, count int
}

func NewPartition(index uint16) *Partition {
	p := &Partition{
		Index:  index,
		events: make([][]byte, minQueueLen),
	}

	return p
}

// Length returns the number of elements currently stored in the queue.
func (p *Partition) Length() int {
	return p.count
}

// resizes the queue to fit exactly twice its current contents
// this can result in shrinking if the queue is less than half-full
func (p *Partition) resize() {
	newBuf := make([][]byte, p.count<<1)

	if p.tail > p.head {
		copy(newBuf, p.events[p.head:p.tail])
	} else {
		n := copy(newBuf, p.events[p.head:])
		copy(newBuf[n:], p.events[:p.tail])
	}

	p.head = 0
	p.tail = p.count
	p.events = newBuf
}

// Add puts an element on the end of the queue.
func (p *Partition) Push(event []byte) {
	if p.count == len(p.events) {
		p.resize()
	}

	p.events[p.tail] = event
	// bitwise modulus
	p.tail = (p.tail + 1) & (len(p.events) - 1)
	p.count++
}

// Peek returns the element at the head of the queue. This call panics
// if the queue is empty.
func (p *Partition) Peek() ([]byte, error) {
	if p.count <= 0 {
		return nil, fmt.Errorf("queue: Peek() called on empty queue")
	}
	return p.events[p.head], nil
}

// Get returns the element at index i in the queue. If the index is
// invalid, the call will panic. This method accepts both positive and
// negative index values. Index 0 refers to the first element, and
// index -1 refers to the last.
func (p *Partition) Get(i int) ([]byte, error) {
	// If indexing backwards, convert to positive index.
	if i < 0 {
		i += p.count
	}
	if i < 0 || i >= p.count {
		return nil, fmt.Errorf("queue: Get() called with index out of range")
	}
	// bitwise modulus
	return p.events[(p.head+i)&(len(p.events)-1)], nil
}

// Remove removes and returns the element from the front of the queue. If the
// queue is empty, the call will panic.
func (p *Partition) Pop() ([]byte, error) {
	if p.count <= 0 {
		return nil, fmt.Errorf("queue: Remove() called on empty queue")
	}
	ret := p.events[p.head]
	p.events[p.head] = nil
	// bitwise modulus
	p.head = (p.head + 1) & (len(p.events) - 1)
	p.count--
	// Resize down if buffer 1/4 full.
	if len(p.events) > minQueueLen && (p.count<<2) == len(p.events) {
		p.resize()
	}
	return ret, nil
}

func (p *Partition) GobDecode(data []byte) error {
	buf := bytes.NewReader(data)

	// decode events
	uint64Bytes := make([]byte, 8)
	_, err := buf.Read(uint64Bytes)
	if err != nil {
		return err
	}
	eventsLength := binary.BigEndian.Uint64(uint64Bytes)
	eventsBytes := make([]byte, eventsLength)
	_, err = buf.Read(eventsBytes)
	if err != nil {
		return err
	}

	eventsReader := bytes.NewReader(eventsBytes)
	decoder := gob.NewDecoder(eventsReader)
	err = decoder.Decode(&p.events)
	if err != nil {
		return err
	}

	// decode Index
	_, err = buf.Read(uint64Bytes)
	if err != nil {
		return err
	}
	p.Index = uint16(binary.BigEndian.Uint64(uint64Bytes))

	// decode head
	_, err = buf.Read(uint64Bytes)
	if err != nil {
		return err
	}
	p.head = int(binary.BigEndian.Uint64(uint64Bytes))

	// decode tail
	_, err = buf.Read(uint64Bytes)
	if err != nil {
		return err
	}
	p.tail = int(binary.BigEndian.Uint64(uint64Bytes))

	// decode count
	_, err = buf.Read(uint64Bytes)
	if err != nil {
		return err
	}
	p.count = int(binary.BigEndian.Uint64(uint64Bytes))

	return nil
}

func (p *Partition) GobEncode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	eventBuffer := bytes.NewBuffer(make([]byte, 0))
	encoder := gob.NewEncoder(eventBuffer)

	// encode events
	err := encoder.Encode(p.events)
	if err != nil {
		return nil, err
	}

	uint64Bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(uint64Bytes, uint64(len(eventBuffer.Bytes())))
	_, err = buf.Write(uint64Bytes)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(eventBuffer.Bytes())
	if err != nil {
		return nil, err
	}

	// encode Index
	binary.BigEndian.PutUint64(uint64Bytes, uint64(p.Index))
	_, err = buf.Write(uint64Bytes)
	if err != nil {
		return nil, err
	}

	// encode head
	binary.BigEndian.PutUint64(uint64Bytes, uint64(p.head))
	_, err = buf.Write(uint64Bytes)
	if err != nil {
		return nil, err
	}

	// encode tail
	binary.BigEndian.PutUint64(uint64Bytes, uint64(p.tail))
	_, err = buf.Write(uint64Bytes)
	if err != nil {
		return nil, err
	}

	// encode count
	binary.BigEndian.PutUint64(uint64Bytes, uint64(p.count))
	_, err = buf.Write(uint64Bytes)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
