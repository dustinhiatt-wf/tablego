/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/4/13
 * Time: 10:57 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"time"
)

const (
	Response		= "response"
	Command			= "command"
)

type ISerializable interface {
	ToBytes()	[]byte
}

type ICellLocation interface {
	Row()		int
	Column()	int
}

type cellLocation struct {
	cellRow			int
	cellColumn		int
}

func (cl *cellLocation) Row() int {
	return cl.cellRow
}

func (cl *cellLocation) Column() int {
	return cl.cellColumn
}

func MakeCellLocation(row, column int) ICellLocation {
	cl := &cellLocation{row, column}
	return cl
}

type IMessage interface {
	SourceTable()			string
	SourceCell()			ICellLocation
	TargetTable()			string
	TargetCell()			ICellLocation
	Operation()				string
	Payload()				[]byte
	MessageId()				string
	Error()					string
	Equal(msg IMessage) 	bool
	GetType()				string
	SetSourceTable(string)
	Timestamp()				int
}

type message struct {
	sourceTable		string
	sourceCell		ICellLocation
	targetTable		string
	targetCell		ICellLocation
	operation		string
	payload			[]byte
	messageId		string
	error			string
	kind			string
	timestamp		int
}

func (m *message) Operation() string {
	return m.operation
}

func (m *message) Payload() []byte {
	return m.payload
}

func (m *message) SourceTable() string {
	return m.sourceTable
}

func (m *message) SourceCell() ICellLocation {
	return m.sourceCell
}

func (m *message) TargetTable() string {
	return m.targetTable
}

func (m *message) TargetCell() ICellLocation {
	return m.targetCell
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

func (m *message) SetSourceTable(value string) {
	m.sourceTable = value
}

func (m *message) Timestamp() int {
	return m.timestamp
}

func (m *message) Equal(msg IMessage) bool {
	if msg.TargetCell().Row() == m.TargetCell().Row() &&
	   msg.TargetCell().Column() == m.TargetCell().Column() &&
	   msg.SourceCell().Row() == m.SourceCell().Row() &&
	   msg.SourceCell().Column() == m.SourceCell().Column() &&
	   msg.TargetTable() == m.TargetTable() &&
	   msg.SourceTable() == m.SourceTable() {
		return true
	}
	return false
}

type ICommand interface {
	IMessage
}

type command struct {
	message
}

func MakeCommand(operation, targetTable, sourceTable string, targetCell, sourceCell ICellLocation, payload []byte) *command {
	message := new(command)
	message.operation = operation
	message.payload = payload
	message.targetCell = targetCell
	message.targetTable = targetTable
	message.sourceCell = sourceCell
	message.sourceTable = sourceTable
	message.messageId = GenUUID()
	message.kind = Command
	message.timestamp = time.Now().Nanosecond()
	return message
}

type IResponse interface {
	IMessage
}

type response struct {
	message
}

func MakeResponse(command ICommand, payload []byte) *response {
	response := new(response)
	response.payload = payload
	response.targetCell = command.SourceCell()
	response.targetTable = command.SourceTable()
	response.sourceCell = command.TargetCell()
	response.sourceTable = command.TargetTable()
	response.messageId = command.MessageId()
	response.operation = command.Operation()
	response.kind = Response
	return response
}

func MakeMessageChannel() chan IMessage {
	return make(chan IMessage)
}

func MakeError(command ICommand, error string) IMessage {
	response := MakeResponse(command, nil)
	response.error = error
	return response
}
