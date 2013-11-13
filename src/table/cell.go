package table

import (
	"encoding/json"
	"log"
)

type ICell interface {
	DisplayValue()				string
	SetValue(value string)
}

type cell struct {
	ISerializable
	CellDisplayValue	string
	Value				string
	LastUpdated			int
	cellChannel			*cellChannel
	observers			*observers
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

func (c *cell) Subscribe(cmd ICommand) {
	c.observers.addObserver(cmd)
}

func (c *cell) send(msg IMessage, ch chan IMessage) {
	ch <- msg
}

func (c *cell) listenToTable() {
	for {
		select {
		case message := <- c.cellChannel.channel.tableToCell:
			if message.GetType() == Response {
				log.Println("ENDING DUE TO RESPONSE")
				continue
			}

			switch message.Operation() {
			case GetCellValue:
				go c.send(MakeResponse(message, c.ToBytes()), c.cellChannel.channel.cellToTable)
			case EditCellValue:
				if message.Timestamp() < c.LastUpdated {
					go c.send(MakeError(message, "Attempted to update with stale value."), c.cellChannel.channel.cellToTable)
					continue
				}

				tblCmd := MakeTableCommandFromJson(message.Payload())
				c.SetValue(tblCmd.Value)
				log.Println("RESPONDING FROM CELL")
				go c.send(MakeResponse(message, c.ToBytes()), c.cellChannel.channel.cellToTable)
			case Subscribe:
				c.Subscribe(message)
				go c.send(MakeResponse(message, c.ToBytes()), c.cellChannel.channel.cellToTable)
			}
		}
	}
}

func MakeCell(row, column int, value string, cc *cellChannel) *cell {
	cell := new(cell)
	cell.Value = value
	cell.CellDisplayValue = value
	cell.cellChannel = cc
	cell.observers = MakeObservers()
	go cell.listenToTable()
	go cell.send(MakeCommand(CellOpened, "", "", MakeCellLocation(row, column), nil, nil), cc.channel.cellToTable)
	return cell
}
