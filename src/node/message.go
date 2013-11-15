package node

import (
	"time"
	"encoding/json"
)

const (
	Response		= "response"
	Command			= "command"
	Error			= "error"
)

type ICoordinates interface {
	Equal(other ICoordinates)			bool
}

type IMessage interface {
	TargetCoordinates()			ICoordinates
	SourceCoordinates()			ICoordinates
	Operation()					string
	Payload()					[]byte
	MessageId()					string
	Error()						string
	Equal(msg IMessage) 		bool
	GetType()					string
	Timestamp()					int
	ToBytes()					[]byte
}

type message struct {
	MTargetCoordinates	ICoordinates
	MSourceCoordinates	ICoordinates
	MOperation			string
	MPayload			[]byte
	MMessageId			string
	MError				string
	MKind				string
	MTimestamp			int
}

func (m *message) ToBytes() []byte {
	res, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return res
}

func MakeMessageFromBytes(bytes []byte) IMessage {
	var m message
	json.Unmarshal(bytes, m)
	return &m
}

func (m *message) Operation() string {
	return m.MOperation
}

func (m *message) Payload() []byte {
	return m.MPayload
}

func (m *message) TargetCoordinates() ICoordinates {
	return m.MTargetCoordinates
}

func (m *message) SourceCoordinates() ICoordinates {
	return m.MSourceCoordinates
}

func (m *message) MessageId() string {
	return m.MMessageId
}

func (m *message) Error() string {
	return m.MError
}

func (m *message) GetType() string {
	return m.MKind
}

func (m *message) Timestamp() int {
	return m.MTimestamp
}

func (m *message) Equal(msg IMessage) bool {
	return m.TargetCoordinates().Equal(msg.TargetCoordinates()) && m.SourceCoordinates().Equal(msg.SourceCoordinates())
}

func makeMessage(operation string, targetCoordinates, sourceCoordinates ICoordinates, payload []byte) *message {
	msg := new(message)
	msg.MTimestamp = time.Now().Nanosecond()
	msg.MOperation = operation
	msg.MTargetCoordinates = targetCoordinates
	msg.MSourceCoordinates = sourceCoordinates
	msg.MPayload = payload
	return msg
}

func MakeMessageChannel() chan IMessage {
	return make(chan IMessage)
}

func MakeCommand(operation string, destination, source ICoordinates, payload []byte) IMessage {
	message := makeMessage(operation, destination, source, payload)
	message.MMessageId = GenUUID()
	message.MKind = Command
	return message
}

func MakeResponse(msg IMessage, payload []byte) IMessage {
	if msg.GetType() != Command {
		panic("Cannot create response from non-command.")
	}
	message := makeMessage(msg.Operation(), msg.SourceCoordinates(), msg.TargetCoordinates(), payload)
	message.MMessageId = msg.MessageId()
	message.MKind = Response
	return message
}

func MakeError(msg IMessage, payload []byte) IMessage {
	message := makeMessage(msg.Operation(), msg.SourceCoordinates(), msg.TargetCoordinates(), payload)
	message.MMessageId = msg.MessageId()
	message.MKind = Error
	return message
}
