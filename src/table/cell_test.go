package table

import (
	"testing"
	"time"
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

func TestCellDoesNotRespondToResponse(t *testing.T) {
	cc := MakeCellChannel()
	MakeCell(1, 1, "", cc)
	<- cc.channel.cellToTable // cell opened
	cmd := MakeCommand("test", "test", "", nil, nil, nil)
	cc.channel.tableToCell <- MakeResponse(cmd, nil)
	go func() {
		<- cc.channel.cellToTable
		t.Error("We got a message from a response.")
	}()
	time.Sleep(20 * time.Millisecond)
}

func TestSerializationDeserializationOfCell(t *testing.T) {
	c := MakeCell(1, 1, "test", MakeCellChannel())
	copy := MakeCellFromBytes(c.ToBytes())
	if c.DisplayValue() != copy.DisplayValue() {
		t.Error("Displayed values do not match.")
	} else if c.Value != copy.Value {
		t.Error("Values do not match.")
	} else if c.LastUpdated != copy.LastUpdated {
		t.Error("Last updated values do not match.")
	}
}

func TestUpdateCellWithStaleData(t *testing.T) {
	cmd := MakeCommand(EditCellValue, "", "", nil, nil, MakeTableCommand("test").ToBytes())
	cc := MakeCellChannel()
	c := MakeCell(1, 1, "test2", cc)
	c.LastUpdated = time.Now().Nanosecond()
	<- cc.channel.cellToTable
	cc.channel.tableToCell <- cmd
	message := <- cc.channel.cellToTable
	if message.Error() == "" {
		t.Error("Error not returned from update with stale data.")
	}
}
