package main

type Message struct {
	FromPeerID string
	Type       MessageType
	Bucket     string
	Key        int
	Value      []byte
}

func (m *Message) IsAliveMessage() bool {
	return m.Type == ALIVE
}

func (m *Message) IsPongMessage() bool {
	return m.Type == PONG
}

type MessageType uint32

const (
	PING MessageType = iota + 1
	PONG
	ELECTION
	ALIVE
	ELECTED
	OK
	REPLICATE_SET
	REPLICATE_DEL
)
