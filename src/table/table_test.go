package table

import (
	"node"
	"testing"
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

	if vr.Values["1"]["1"] != "test" {
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
	if vr.Values["1"]["1"] != "test" {
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
	if vr.Values["1"]["1"] != "test" {
		t.Error("Value range not returned correctly.")
	} else if vr.Values["1"]["2"] != "test2" {
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

/*

func TestEditValueSubscribe(t *testing.T) {
	table := MakeTable("test", nil, nil)
	cellCh := MakeValueChannel()
	table.EditTableValue(1, 1, "test", cellCh)
	message := <- cellCh
	cell := message.cell
	ch := MakeValueChannel()
	cell.Subscribe(ch)
	go func () {
		for {
			select {
			case message := <- ch:
				if message.cell.value != "test2" {
					t.Error("Subscribed not sending correct values.")
				}
				return
			}
		}
	}()
	table.EditTableValue(1, 1, "test2", MakeValueChannel())
	<- ch
}

func TestGetRangeFromTable(t *testing.T) {
	table := MakeTable("test", nil, nil)
	table.EditTableValue(1, 1, "test", MakeValueChannel())
	ch := MakeValueChannel()
	table.GetRangeByRowAndColumn(0, 2, 0, 2, ch)
	message := <- ch
	tr := message.tableRange
	if table.cells[1][1] != tr.cells[1][1] {
		t.Error("Get table range did not return correct values.")
	}
}

func TestSum(t *testing.T) {
	table := MakeTable("test", MakeOrchestrator(), nil)
	table.EditTableValue(1, 1, "7", MakeValueChannel())
	table.EditTableValue(2, 1, "3", MakeValueChannel())
	ch := MakeValueChannel()
	table.EditTableValue(3, 1, "=sum(A1:C3)", ch)
	message := <- ch
	sumCell := message.cell
	if sumCell.DisplayValue != "10" {
		t.Error("Single column not summed correctly.")
	}
}

func TestSumWithFloat(t *testing.T) {
	table := MakeTable("test", nil, nil)
	table.EditTableValue(1, 1, "7.2", MakeValueChannel())
	table.EditTableValue(2, 1, "2.7", MakeValueChannel())
	ch := MakeValueChannel()
	table.EditTableValue(3, 1, "=sum(A1:C3)", ch)
	message := <- ch
	sumCell := message.cell
	if sumCell.DisplayValue != "9.9" {
		t.Error("Single column float not summed correctly.")
	}
}

func TestSumOverRange(t *testing.T) {
	table := MakeTable("test", nil, nil)
	table.EditTableValue(1, 1, "5", MakeValueChannel())
	table.EditTableValue(2, 2, "5", MakeValueChannel())
	ch := MakeValueChannel()
	table.EditTableValue(5, 0, "=sum(A1:C3)", ch)
	message := <- ch
	sumCell := message.cell
	if sumCell.DisplayValue != "10" {
		t.Error("Range summation did not work correctly.")
	}
}

func TestSumUpdatesWhenCellUpdates(t *testing.T) {
	table := MakeTable("test", nil, nil)
	table.EditTableValue(1, 1, "5", MakeValueChannel())
	table.EditTableValue(2, 2, "5", MakeValueChannel())
	ch := MakeValueChannel()
	table.EditTableValue(5, 0, "=sum(A1:C3)", ch)
	message := <- ch
	sumCell := message.cell
	ch = MakeValueChannel()
	sumCell.Subscribe(ch)
	table.EditTableValue(1, 1, "7", MakeValueChannel())
	<- ch
	if sumCell.DisplayValue != "12" {
		t.Error("Summation cell not updated correctly.")
	}
}

func TestSumUpdatesWhenEmptyCellUpdates(t *testing.T) {
	table := MakeTable("test", nil, nil)
	table.EditTableValue(1, 1, "5", MakeValueChannel())
	ch := MakeValueChannel()
	table.EditTableValue(5, 0, "=sum(A1:C3)", ch)
	message := <- ch
	sumCell := message.cell
	ch = MakeValueChannel()
	sumCell.Subscribe(ch)
	table.EditTableValue(2, 2, "5", MakeValueChannel())
	<- ch
	if sumCell.DisplayValue != "10" {
		t.Error("Sum did not update with empty cell correctly.")
	}
}

func TestSumUpdatesCascadeAcrossSums(t *testing.T) {
	table := MakeTable("test", nil, nil)
	table.EditTableValue(1, 4, "5", MakeValueChannel())
	table.EditTableValue(1, 1, "=sum(E1:F2)", MakeValueChannel())
	table.EditTableValue(2, 1, "5", MakeValueChannel())
	ch := MakeValueChannel()
	table.EditTableValue(5, 0, "=sum(A1:C3)", ch)
	message := <- ch
	sumCell := message.cell
	ch = MakeValueChannel()
	sumCell.Subscribe(ch)
	if sumCell.DisplayValue != "10" {
		t.Error("Initial summation not calculated correctly.")
	}
	table.EditTableValue(1, 4, "10", MakeValueChannel())
	<- ch // final value of 15
	if sumCell.DisplayValue != "15" {
		t.Error("Final summation not calculated correctly.")
	}
}

func TestTableNotifiesSubscribersWhenCellsChange(t *testing.T) {
	ch := MakeValueChannel()
	chTwo := MakeValueChannel()
	table := MakeTable("test", nil, nil)
	table.Subscribe(ch)
	table.EditTableValue(1, 1, "1", chTwo)
	message := <- chTwo
	cell := message.cell
	message = <- ch
	if cell != message.cell {
		t.Error("Table not correctly notifying listeners.")
	}
}

func TestFormulaRangeChange(t *testing.T) {
	ch := MakeValueChannel()
	sumCh := MakeValueChannel()
	table := MakeTable("test", nil, nil)
	table.EditTableValue(1, 1, "5", ch)
	message := <- ch
	cell := message.cell
	ch = MakeValueChannel()
	cell.Subscribe(ch)
	table.EditTableValue(5, 0, "=sum(A1:C3)", sumCh)
	message = <- sumCh
	sumCell := message.cell
	sumCh = MakeValueChannel()
	sumCell.Subscribe(sumCh)
	if sumCell.DisplayValue != "5" {
		t.Error("Sum cell not calculated correctly.")
	}
	table.EditTableValue(5, 0, "=sum(E1:G3)", MakeValueChannel())
	<- sumCh
	if sumCell.DisplayValue != "0" {
		t.Error("Sum cell not calculated correctly after sum range update.")
	}
	table.EditTableValue(1, 1, "10", MakeValueChannel())
	<- ch
	if sumCell.DisplayValue != "0" {
		t.Error("Sum cell recalculated using stale range.")
	}
	table.EditTableValue(1, 5, "5", MakeValueChannel())
	<- sumCh
	if sumCell.DisplayValue != "5" {
		t.Error("Formula did not update correctly with new range.")
	}
}

/*
This will deadlock if notifications from table aren't sent correctly

func TestTableNotifiesWhenCellsUpdated(t *testing.T) {
	ch := MakeValueChannel()
	table := MakeTable("test", nil, nil)
	table.Subscribe(ch)
	table.EditTableValue(1, 1, "1", MakeValueChannel())
	table.EditTableValue(2, 1, "1", MakeValueChannel())
	table.EditTableValue(5, 0, "=sum(A1:C3)", MakeValueChannel())
	<- ch
	<- ch
	<- ch
}

func TestCascadingUpdatesSendsNotificationsThroughTable(t *testing.T) {
	ch := MakeValueChannel()
	table := MakeTable("test", nil, nil)
	table.Subscribe(ch)
	table.EditTableValue(1, 5, "5", MakeValueChannel())
	table.EditTableValue(1, 1, "=sum(E1:G3)", MakeValueChannel())
	table.EditTableValue(2, 1, "5", MakeValueChannel())
	sumCh := MakeValueChannel()
	table.EditTableValue(5, 0, "=sum(A1:C3)", sumCh)
	message := <- sumCh
	sumCell := message.cell
	<- ch
	<- ch
	<- ch
	message = <- ch
	if message.cell != sumCell {
		t.Error("Table not correctly handing subscriptions.")
	}
	table.EditTableValue(1, 5, "10", MakeValueChannel())
	<- ch // constant updated
	<- ch // first sum updated
	message = <- ch // second sum updated
	if message.cell != sumCell {
		t.Error("Cascading formulas not notifying table.")
	}
}

func TestCrossTableFormula(t *testing.T) {
	o := MakeOrchestrator()
	ch := MakeValueChannel()
	o.GetTableById("test1", ch)
	message := <- ch
	tableOne := message.table
	ch = MakeValueChannel()
	o.GetTableById("test2", ch)
	message = <- ch
	tableTwo := message.table
	tableOne.EditTableValue(1, 1, "5", MakeValueChannel())
	ch = MakeValueChannel()
	tableTwo.EditTableValue(5, 0, "=sum(test1:A1:C3)", ch)
	message = <- ch
	cell := message.cell
	if cell.DisplayValue != "5" {
		t.Error("Cross table formula not calculated correctly.")
	}
}

func TestCrossTableFormulaWithUpdates(t *testing.T) {
	o := MakeOrchestrator()
	ch := MakeValueChannel()
	o.GetTableById("test1", ch)
	message := <- ch
	tableOne := message.table
	ch = MakeValueChannel()
	o.GetTableById("test2", ch)
	message = <- ch
	tableTwo := message.table
	ch = MakeValueChannel()
	tableOne.EditTableValue(1, 1, "5", ch)
	message = <- ch
	constCell := message.cell
	ch = MakeValueChannel()
	formulaListener := MakeValueChannel()
	tableTwo.EditTableValue(5, 0, "=sum(test1:A1:C3)", ch)
	message = <- ch
	cell := message.cell
	cell.Subscribe(formulaListener)
	cellOneListener := MakeValueChannel()
	constCell.Subscribe(cellOneListener)
	tableOne.EditTableValue(1, 1, "10", ch)
	<- cellOneListener
	<- formulaListener
	if cell.DisplayValue != "10" {
		t.Error("Cross table formula not updated correctly.")
	}
}*/
