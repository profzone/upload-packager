package consensus

import (
	"bytes"
	"encoding/gob"
)

type CreatePeerBody struct {
	PeerID   string `json:"peerID"`
	PeerAddr string `json:"peerAddr"`
}

type Marshaler interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type LogEntryEvent struct {
	Key   string
	Value string
}

func (l LogEntryEvent) Marshal() (result []byte, err error) {
	buf := bytes.NewBuffer(result)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(l)
	return buf.Bytes(), err
}

func (l *LogEntryEvent) Unmarshal(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(l)
	return
}
