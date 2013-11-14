package table

import (
	"encoding/json"
	"node"
)

type ICell interface {
	DisplayValue()				string
	SetValue(value string)
}

type cell struct {
	ISerializable
	ICell
	node.INode
	node.ICommunicationHandler
	node.INodeFactory
	CellDisplayValue				string
	Value							string
	LastUpdated						int
	observers						*observers
	pendingRequests					map[string]chan node.IMessage
}

func (c *cell) ToBytes() []byte {
	res, err := json.Marshal(c)
	if err != nil {
		return nil
	}

	return res
}

func MakeCellFromBytes(bytes []byte) *cell {
	var m cell
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return nil
	}
	return &m
}

func (c *cell) DisplayValue() string {
	return c.CellDisplayValue
}

func (c *cell) SetValue(value string) {
	if value == c.Value {
		return
	}
	c.Value = value
	c.CellDisplayValue = value
	go c.observers.notifyObservers(CellUpdated, c.cellChannel.channel.cellToTable, c.ToBytes())
}

func (c *cell) onMessageFromParent(msg IMessage) {
	if msg.GetType() == node.Response {

	} else if msg.GetType() == node.Command {
		switch message.Operation() {
		case GetCellValue:
			resp := node.MakeResponse(message, c.ToBytes())
			go c.INode.send(c.INode.Parent().ChildToParent(), resp)
		case EditCellValue:
			if message.Timestamp() < c.LastUpdated {
				err := node.MakeError(message, "You have attempted a stale update.")
				go c.send(c.INode.Parent().ChildToParent(), err)
				continue
			}

			tblCmd := MakeTableCommandFromJson(message.Payload())
			c.SetValue(tblCmd.Value)
			resp := node.MakeResponse(message, c.ToBytes())
			go c.send(c.INode.Parent().ChildToParent(), resp)
		}
	} else {
		//TODO: log error
	}
}

func (c *cell) onMessageFromChild(msg IMessage) {
	panic("Cells should not receive messages from children")
}

func (c *cell) GetChild(coords ICoordinates) IChild {
	//TODO: so we can handle embedded formulas, they should be children
	panic("Cells have no children currently.")
}

func (t *table) makeChildNode(parentChannel IChild, childCoordinates ICoordinates) INode {
	panic("Cells can't create children.")
}

func (c *cell) Subscribe(msg node.IMessage) {
	c.observers.addObserver(msg)
}

func MakeCell(parentChannel IChannel, coordinates, parentCoordinates ICoordinates, value string) *cell {
	cell := new(cell)
	cell.Value = value
	cell.CellDisplayValue = value
	cell.observers = MakeObservers()
	cell.pendingRequests = make(map[string]chan node.IMessage)
	cell.INode = node.MakeNode(parentChannel, coordinates, parentCoordinates, cell, cell)
	return cell
}
