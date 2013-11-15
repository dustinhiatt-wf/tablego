package table

import (
	"encoding/json"
	"node"
)

type ICell interface {
	DisplayValue()				string
	SetValue(value string)
	GetCellValue()				*cellValue
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

type cellValue struct {
	CellDisplayValue				string
	Value							string
	LastUpdated						int
}

func makeCellValue(cellDisplayValue, value string, lastUpdated int) *cellValue {
	cv := new(cellValue)
	cv.CellDisplayValue = cellDisplayValue
	cv.Value = value
	cv.LastUpdated = lastUpdated
	return cv
}

func (c *cell) GetCellValue() *cellValue {
	return makeCellValue(c.DisplayValue(), c.Value, c.LastUpdated)
}

func (c *cell) ToBytes() []byte {
	cv := c.GetCellValue()
	res, err := json.Marshal(cv)
	if err != nil {
		return nil
	}

	return res
}

func MakeCellFromBytes(bytes []byte) *cell {
	var m cellValue
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return nil
	}
	cv := &m
	cell := new(cell)
	cell.CellDisplayValue = cv.CellDisplayValue
	cell.Value = cv.Value
	cell.LastUpdated = cv.LastUpdated
	return cell
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
	go c.observers.notifyObservers(CellUpdated, c.INode.Parent().ChildToParent(), c.ToBytes())
}

func (c *cell) OnMessageFromParent(msg node.IMessage) {
	if msg.GetType() == node.Response {

	} else if msg.GetType() == node.Command {
		switch msg.Operation() {
		case GetCellValue:
			resp := node.MakeResponse(msg, c.ToBytes())
			go c.INode.Send(c.INode.Parent().ChildToParent(), resp)
		case EditCellValue:
			if msg.Timestamp() < c.LastUpdated {
				err := node.MakeError(msg, nil)
				go c.Send(c.INode.Parent().ChildToParent(), err)
				return
			}

			tblCmd := MakeTableCommandFromJson(msg.Payload())
			c.SetValue(tblCmd.Value)
			resp := node.MakeResponse(msg, c.ToBytes())
			go c.Send(c.INode.Parent().ChildToParent(), resp)
		default:
			resp := node.MakeResponse(msg, nil)
			c.INode.Send(c.INode.Parent().ChildToParent(), resp)
		}

	} else {
		//TODO: log error
	}
}

func (c *cell) OnMessageFromChild(msg node.IMessage) {
	panic("Cells should not receive messages from children")
}

func (c *cell) GetChild(coords node.ICoordinates) node.IChild {
	//TODO: so we can handle embedded formulas, they should be children
	panic("Cells have no children currently.")
}

func (c *cell) makeChildNode(parentChannel node.IChild, childCoordinates node.ICoordinates) node.INode {
	panic("Cells can't create children.")
}

func (c *cell) Subscribe(msg node.IMessage) {
	c.observers.addObserver(msg)
}

func MakeCell(parentChannel node.IChannel, coordinates, parentCoordinates node.ICoordinates, value string) *cell {
	cell := new(cell)
	cell.Value = value
	cell.CellDisplayValue = value
	cell.observers = MakeObservers()
	cell.pendingRequests = make(map[string]chan node.IMessage)
	cell.INode = node.MakeNode(parentChannel, coordinates, parentCoordinates, cell, cell)
	return cell
}
