package table


import (
	"testing"
	"node"
	"time"
)

func TestSubscribeToCell(t *testing.T) {
	chch := node.MakeIChild()
	MakeCell(chch.Channel(), MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("test", nil), "")
	<- chch.Channel().ChildToParent() //cell initialized
	chch.Channel().ParentToChild() <- node.MakeCommand(Subscribe, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("test", nil), makeSubscribePayload("", -1, -1, false).ToBytes())
	<- chch.Channel().ChildToParent() // subscribed
	chch.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())
	var message node.IMessage
	msg := <- chch.Channel().ChildToParent() // updated
	if msg.Operation() == CellUpdated {
		message = msg
	}
	msg = <- chch.Channel().ChildToParent() //subscribe message
	if msg.Operation() == CellUpdated {
		message = msg
	}
	if message.Operation() != CellUpdated {
		t.Error("Updated message not received")
	}
	cell := MakeCellFromBytes(message.Payload())
	if cell.DisplayValue() != "test" {
		t.Error("Updated cell does not have correct value.")
	}
}

func TestUnsubscribeToCell(t *testing.T) {
	chch := node.MakeIChild()
	MakeCell(chch.Channel(), MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("test", nil), "")
	<- chch.Channel().ChildToParent() //cell initialized
	chch.Channel().ParentToChild() <- node.MakeCommand(Subscribe, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("test", nil), makeSubscribePayload("", -1, -1, false).ToBytes())
	<- chch.Channel().ChildToParent() // subscribed
	chch.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())
	<- chch.Channel().ChildToParent() // updated
	<- chch.Channel().ChildToParent() // subscribed updated

	chch.Channel().ParentToChild() <- node.MakeCommand(Unsubscribe, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("test", nil), makeSubscribePayload("", -1, -1, false).ToBytes())
	<- chch.Channel().ChildToParent() // unsubscribed
	chch.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test2").ToBytes())
	<- chch.Channel().ChildToParent() //updated

	go func () {
		<- chch.Channel().ChildToParent() //this shouldn't happen
		t.Error("Subscribe update received.")
	}()

	time.Sleep(20 * time.Millisecond) // don't like this but don't have time for a better option
}

func TestDoesNotUpdateWithIdenticalValue(t *testing.T) {
	chch := node.MakeIChild()
	MakeCell(chch.Channel(), MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("test", nil), "")
	<- chch.Channel().ChildToParent() //cell initialized
	chch.Channel().ParentToChild() <- node.MakeCommand(Subscribe, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("test", nil), makeSubscribePayload("", -1, -1, false).ToBytes())
	<- chch.Channel().ChildToParent() // subscribed
	chch.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())
	<- chch.Channel().ChildToParent() // updated
	<- chch.Channel().ChildToParent() // subscribed updated

	chch.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())
	<- chch.Channel().ChildToParent() //updated

	go func () {
		<- chch.Channel().ChildToParent() //this shouldn't happen
		t.Error("Subscribe update received.")
	}()

	time.Sleep(20 * time.Millisecond) // don't like this but don't have time for a better option
}

func TestSerializationDeserializationOfCell(t *testing.T) {
	chch := node.MakeIChild()
	c := MakeCell(chch.Channel(), MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("test", nil), "test")
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
	chch := node.MakeIChild()
	MakeCell(chch.Channel(), MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("test", nil), "")
	<- chch.Channel().ChildToParent() //cell initialized
	chch.Channel().ParentToChild() <- node.MakeCommand(Subscribe, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("test", nil), makeSubscribePayload("", -1, -1, false).ToBytes())
	<- chch.Channel().ChildToParent() // subscribed
	cmdOld := node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())
	cmdNew := node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test2").ToBytes())
	chch.Channel().ParentToChild() <- cmdNew
	<- chch.Channel().ChildToParent()
	<- chch.Channel().ChildToParent() // updated

	chch.Channel().ParentToChild() <- cmdOld

	msg := <- chch.Channel().ChildToParent()
	if msg.GetType() != node.Error {
		t.Error("Stale data did not throw error.")
	}

	go func() {
		<- chch.Channel().ChildToParent()
		t.Error("Notified on stale update.")
	}()

	time.Sleep(20 * time.Millisecond)
}

func TestParseFormulaIsNotAFormula(t *testing.T) {
	chch := node.MakeIChild()
	c := MakeCell(chch.Channel(), MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("test", nil), "")
	<- chch.Channel().ChildToParent() //cell initialized

	c.parseValue("test")
	if c.isFormula == true {
		t.Error("Formula not parsed correctly.")
	}
}

func TestCellSubscribesWhenFormulaSet(t *testing.T) {
	chch := node.MakeIChild()
	vr := new(valuerange)
	vr.Values = make(map[string]map[string]string)
	vr.Values["1"] = make(map[string]string)
	vr.Values["2"] = make(map[string]string)
	vr.Values["1"]["4"] = "5.5"
	vr.Values["2"]["532"] = "7"
	vr.Values["2"]["0"] = "test"
	var msg node.IMessage
	go func () {
		msg = <- chch.Channel().ChildToParent()
		chch.Channel().ParentToChild() <- node.MakeResponse(msg, vr.ToBytes())


	}()
	MakeCell(chch.Channel(), MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("test", nil), "=sum(A1:C3)")
	if msg.Operation() != SubscribeToRange {
		t.Error("Cell did not subscribe to range of interest.")
	}
}

/*
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
*/
