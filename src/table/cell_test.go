package table

import (
	"testing"
)

func TestUpdateValueNotifyObservers(t *testing.T) {
	cc := MakeCellChannel()
	cell := MakeCell(1, 1, "", cc)
	<- cc.channel.cellToTable // cell Opened
	cc.channel.tableToCell <- MakeCommand(Subscribe, "", "test", MakeCellLocation(1, 1), MakeCellLocation(2, 2), nil)
	cell.SetValue("test")
	message := <- cc.channel.cellToTable // subscribe notification
	message = <- cc.channel.cellToTable // update notification
	if message.Operation() != CellUpdated {
		t.Error("Subscribe message has wrong operation type.")
	} else if message.TargetTable() != "test" {
		t.Error("Subscribe message has wrong target table.")
	} else if message.TargetCell().Row() != 2 {
		t.Error("Subscribe message has wrong target row.")
	} else if message.TargetCell().Column() != 2 {
		t.Error("Subscribe message has wrong target column.")
	}
	c := MakeCellFromBytes(message.Payload())
	if c.Value != "test" {
		t.Error("Subscribe message has wrong updated value.")
	}
}
