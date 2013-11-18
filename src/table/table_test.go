package table

import (
	"node"
	"testing"
	"time"
)

func TestCreateTable(t *testing.T) {
	child := node.MakeIChild()
	table := MakeTable(child.Channel(), MakeCoordinates("test", nil), MakeCoordinates("", nil))
	<-child.Channel().ChildToParent() // table created
	if table.children == nil {
		t.Error("Table not created correctly.")
	}
}

func TestTableCreatesCell(t *testing.T) {
	child := node.MakeIChild()
	table := MakeTable(child.Channel(), MakeCoordinates("test", nil), MakeCoordinates("", nil))
	<-child.Channel().ChildToParent() // table initialized
	child.Channel().ParentToChild() <- node.MakeCommand("test", MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), nil)
	message := <-child.Channel().ChildToParent()
	if message.Operation() != "test" {
		t.Error("Cell not created correctly.")
	} else if len(table.children) != 1 {
		t.Error("Cell not cached correctly.")
	}
}

func TestForwardsToCells(t *testing.T) {
	child := node.MakeIChild()
	table := MakeTable(child.Channel(), MakeCoordinates("test", nil), MakeCoordinates("", nil))
	<-child.Channel().ChildToParent() //table initialized
	child.Channel().ParentToChild() <- node.MakeCommand("test", MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), nil)
	<-child.Channel().ChildToParent() // cell initialized
	cellChild := node.MakeIChild()
	table.children[1][2] = cellChild
	table.children[1][1].Channel().ChildToParent() <- node.MakeCommand("test", MakeCoordinates("test", MakeCellLocation(1, 2)), MakeCoordinates("test", MakeCellLocation(1, 1)), nil)
	msg := <-cellChild.Channel().ParentToChild()
	if msg.Operation() != "test" {
		t.Error("Table not forwarding to cells correctly.")
	}
}

func TestGetValueRange(t *testing.T) {
	child := node.MakeIChild()
	MakeTable(child.Channel(), MakeCoordinates("test", nil), MakeCoordinates("", nil))
	<-child.Channel().ChildToParent() //table initialized
	child.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())
	<-child.Channel().ChildToParent() // cell's value set
	cr := MakeRange("A1:C3")

	child.Channel().ParentToChild() <- node.MakeCommand(GetValueRange, MakeCoordinates("test", nil), MakeCoordinates("", nil), cr.ToBytes())
	message := <-child.Channel().ChildToParent()

	vr := MakeValueRangeFromBytes(message.Payload())

	if vr.Values["1"]["1"].CellDisplayValue != "test" {
		t.Error("Value range not returned correctly.")
	}
}

func TestGetMultipleValueRange(t *testing.T) {
	child := node.MakeIChild()
	MakeTable(child.Channel(), MakeCoordinates("test", nil), MakeCoordinates("", nil))
	<-child.Channel().ChildToParent() // cell's value set
	for i := 0; i < 100; i++ {
		child.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(i, 0)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())
	}
	for i := 0; i < 100; i++ {
		<-child.Channel().ChildToParent()
	}
	cr := new(cellrange)
	cr.StartRow = 0
	cr.StopRow = 101
	cr.StartColumn = 0
	cr.StopColumn = 1
	child.Channel().ParentToChild() <- node.MakeCommand(GetValueRange, MakeCoordinates("test", nil), MakeCoordinates("", nil), cr.ToBytes())
	message := <-child.Channel().ChildToParent()

	vr := MakeValueRangeFromBytes(message.Payload())
	if len(vr.Values) != 100 {
		t.Error("Wrong value range size")
	}
}

func TestSubscribeToRange(t *testing.T) {
	child := node.MakeIChild()
	MakeTable(child.Channel(), MakeCoordinates("test", nil), MakeCoordinates("", nil))
	<-child.Channel().ChildToParent() //table initialized
	child.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())
	<-child.Channel().ChildToParent() // cell's value set
	cr := MakeRange("A1:C3")

	child.Channel().ParentToChild() <- node.MakeCommand(SubscribeToRange, MakeCoordinates("test", nil), MakeCoordinates("", nil), cr.ToBytes())
	msg := <-child.Channel().ChildToParent()
	vr := MakeValueRangeFromBytes(msg.Payload())
	if vr.Values["1"]["1"].CellDisplayValue != "test" {
		t.Error("Value range not returned from subscribe correctly.")
	}
}

func TestSubscribeToRangeWithMultipleValues(t *testing.T) {
	child := node.MakeIChild()
	MakeTable(child.Channel(), MakeCoordinates("test", nil), MakeCoordinates("", nil))
	<-child.Channel().ChildToParent() //table initialized
	child.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())
	child.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 2)), MakeCoordinates("", nil), MakeTableCommand("test2").ToBytes())
	<-child.Channel().ChildToParent()
	<-child.Channel().ChildToParent() // cell's value set
	cr := MakeRange("A1:C3")

	child.Channel().ParentToChild() <- node.MakeCommand(SubscribeToRange, MakeCoordinates("test", nil), MakeCoordinates("", nil), cr.ToBytes())
	msg := <-child.Channel().ChildToParent()
	vr := MakeValueRangeFromBytes(msg.Payload())
	if vr.Values["1"]["1"].CellDisplayValue != "test" {
		t.Error("Value range not returned correctly.")
	} else if vr.Values["1"]["2"].CellDisplayValue != "test2" {
		t.Error("Value range not returned correctly.")
	}
}

func TestSubscribeToRangeNotifiesOnBlankUpdate(t *testing.T) {
	child := node.MakeIChild()
	MakeTable(child.Channel(), MakeCoordinates("test", nil), MakeCoordinates("", nil))
	<-child.Channel().ChildToParent() //table initialized
	cr := MakeRange("A1:C3")

	child.Channel().ParentToChild() <- node.MakeCommand(SubscribeToRange, MakeCoordinates("test", nil), MakeCoordinates("", nil), cr.ToBytes())
	<-child.Channel().ChildToParent()

	child.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())

	<- child.Channel().ChildToParent() // one of these lets me know the cell was edited and the other lets me know that we
	<- child.Channel().ChildToParent() // received a subscription notification
}

func TestUnsubscribeToRange(t *testing.T) {
	child := node.MakeIChild()
	MakeTable(child.Channel(), MakeCoordinates("test", nil), MakeCoordinates("", nil))
	<-child.Channel().ChildToParent() //table initialized
	cr := MakeRange("A1:C3")

	child.Channel().ParentToChild() <- node.MakeCommand(SubscribeToRange, MakeCoordinates("test", nil), MakeCoordinates("", nil), cr.ToBytes())
	<-child.Channel().ChildToParent()

	child.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())

	<- child.Channel().ChildToParent() // one of these lets me know the cell was edited and the other lets me know that we
	<- child.Channel().ChildToParent() // received a subscription notification

	child.Channel().ParentToChild() <- node.MakeCommand(UnsubscribeToRange, MakeCoordinates("test", nil), MakeCoordinates("", nil), cr.ToBytes())
	<- child.Channel().ChildToParent() // unsubscribed
	child.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test2").ToBytes())
	<- child.Channel().ChildToParent() // updated

	go func () {
		<- child.Channel().ChildToParent()
		t.Error("Unsubscription to range not successful.")
	}()

	time.Sleep(20 * time.Millisecond)
}

func TestFormulaUpdateTest(t *testing.T) {
	child := node.MakeIChild()
	MakeTable(child.Channel(), MakeCoordinates("test", nil), MakeCoordinates("", nil))
	<-child.Channel().ChildToParent() //table initialized


	child.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("=sum(A1:C3)").ToBytes())
	msg := <- child.Channel().ChildToParent()// we now have a cell
	c := MakeCellFromBytes(msg.Payload())

	if c.DisplayValue() != "0" {
		t.Error("sum not calculated correctly.")
	}

	child.Channel().ParentToChild() <- node.MakeCommand(Subscribe, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), nil)
	msg = <- child.Channel().ChildToParent() // we are now subscribed

	child.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 2)), MakeCoordinates("", nil), MakeTableCommand("1").ToBytes())
	var message node.IMessage

	msg = <- child.Channel().ChildToParent()
	if msg.Operation() == CellUpdated {
		message = msg
	}

	msg = <- child.Channel().ChildToParent()
	if msg.Operation() == CellUpdated {
		message = msg
	}

	c = MakeCellFromBytes(message.Payload())

	if c.DisplayValue() != "1" {
		t.Error("Formula not updated correctly.")
	}
}

func BenchmarkAddAndGetCells(b *testing.B) {
	child := node.MakeIChild()
	MakeTable(child.Channel(), MakeCoordinates("test", nil), MakeCoordinates("", nil))
	<-child.Channel().ChildToParent() // cell's value set
	for i := 0; i < b.N; i++ {
		child.Channel().ParentToChild() <- node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(i, 0)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())
	}
	for i := 0; i < b.N; i++ {
		<-child.Channel().ChildToParent()
	}
	cr := new(cellrange)
	cr.StartRow = 0
	cr.StopRow = b.N + 1
	cr.StartColumn = 0
	cr.StopColumn = 1
	b.ResetTimer()
	child.Channel().ParentToChild() <- node.MakeCommand(GetValueRange, MakeCoordinates("test", nil), MakeCoordinates("", nil), cr.ToBytes())
	message := <-child.Channel().ChildToParent()

	vr := MakeValueRangeFromBytes(message.Payload())
	if len(vr.Values) != b.N {
		b.Error("Wrong value range size")
	}
}
