package node

import (
	"time"
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
}

type message struct {
	targetCoordinates	ICoordinates
	sourceCoordinates	ICoordinates
	operation			string
	payload				[]byte
	messageId			string
	error				string
	kind				string
	timestamp			int
}

func (m *message) Operation() string {
	return m.operation
}

func (m *message) Payload() []byte {
	return m.payload
}

func (m *message) TargetCoordinates() ICoordinates {
	return m.targetCoordinates
}

func (m *message) SourceCoordinates() ICoordinates {
	return m.sourceCoordinates
}

func (m *message) MessageId() string {
	return m.messageId
}

func (m *message) Error() string {
	return m.error
}

func (m *message) GetType() string {
	return m.kind
}

func (m *message) Timestamp() int {
	return m.timestamp
}

func (m *message) Equal(msg IMessage) bool {
	return m.TargetCoordinates().Equal(msg.TargetCoordinates()) && m.SourceCoordinates().Equal(msg.SourceCoordinates())
}

func makeMessage(operation string, targetCoordinates, sourceCoordinates ICoordinates, payload []byte) *message {
	msg := new(message)
	msg.timestamp = time.Now().Nanosecond()
	msg.operation = operation
	msg.targetCoordinates = targetCoordinates
	msg.sourceCoordinates = sourceCoordinates
	msg.payload = payload
	return msg
}

func MakeMessageChannel() chan IMessage {
	return make(chan IMessage)
}

func MakeCommand(operation string, destination, source ICoordinates, payload []byte) IMessage {
	message := makeMessage(operation, destination, source, payload)
	message.messageId = GenUUID()
	message.kind = Command
	return message
}

func MakeResponse(msg IMessage, payload []byte) IMessage {
	if msg.GetType() != Command {
		panic("Cannot create response from non-command.")
	}
	message := makeMessage(msg.Operation(), msg.SourceCoordinates(), msg.TargetCoordinates(), payload)
	message.messageId = msg.MessageId()
	message.kind = Response
	return message
}

func MakeError(msg IMessage, payload []byte) IMessage {
	message := makeMessage(msg.Operation(), msg.SourceCoordinates(), msg.TargetCoordinates(), payload)
	message.messageId = msg.MessageId()
	message.kind = Error
	return message
}
