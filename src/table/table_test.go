/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/4/13
 * Time: 11:25 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"testing"
)

func TestCreateTable(t *testing.T) {
	table := MakeTable("test", nil)
	if table.cells == nil {
		t.Error("Table cells not initialized.")
	}
}

func TestEditValueAtPosition(t *testing.T) {
	table := MakeTable("test", nil)
	table.EditTableValue(1, 1, "test")
	if _, ok :=table.cells[1][1]; !ok {
		t.Error("Channel package not set.")
	}
}

func TestGetValueExists(t *testing.T) {
	table := MakeTable("test", nil)
	table.EditTableValue(1, 1, "test")
	cell := table.GetValueAt(1, 1)
	if cell.value != "test" {
		t.Error("Value not retrieved correctly.")
	}
}

func TestGetValueDoesNotExist(t *testing.T) {
	table := MakeTable("test", nil)
	cell := table.GetValueAt(1, 1)
	if cell != nil {
		t.Error("No cell should have been received.")
	}
}

func TestEditValueSubscribe(t *testing.T) {
	table := MakeTable("test", nil)
	table.EditTableValue(1, 1, "test")
	ch := MakeValueChannel()
	table.Subscribe(1, 1, ch)
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
	table.EditTableValue(1, 1, "test2")
	<- ch
}

func TestSubscribeNonValue(t *testing.T) {
	table := MakeTable("test", nil)
	ch := MakeValueChannel()
	table.Subscribe(1, 1, ch)
	table.EditTableValue(1, 1, "1")
	message := <- ch
	if message.cell.DisplayValue != "1" {
		t.Error("Subscribed to non-value is not working properly.")
	}
}

func TestGetRangeFromTable(t *testing.T) {
	table := MakeTable("test", nil)
	table.EditTableValue(1, 1, "test")
	tr := table.GetRangeByRowAndColumn(0, 2, 0, 2)
	if table.cells[1][1] != tr.cells[1][1] {
		t.Error("Get table range did not return correct values.")
	}
}

func TestSum(t *testing.T) {
	table := MakeTable("test", nil)
	table.EditTableValue(1, 1, "7")
	table.EditTableValue(2, 1, "3")
	sumCell := table.EditTableValue(3, 1, "=sum(A1:C3)")
	if sumCell.DisplayValue != "10" {
		t.Error("Single column not summed correctly.")
	}
}

func TestSumWithFloat(t *testing.T) {
	table := MakeTable("test", nil)
	table.EditTableValue(1, 1, "7.2")
	table.EditTableValue(2, 1, "2.7")
	sumCell := table.EditTableValue(3, 1, "=sum(A1:C3)")
	if sumCell.DisplayValue != "9.9" {
		t.Error("Single column float not summed correctly.")
	}
}

func TestSumOverRange(t *testing.T) {
	table := MakeTable("test", nil)
	table.EditTableValue(1, 1, "5")
	table.EditTableValue(2, 2, "5")
	sumCell := table.EditTableValue(5, 0, "=sum(A1:C3)")
	if sumCell.DisplayValue != "10" {
		t.Error("Range summation did not work correctly.")
	}
}

func TestSumUpdatesWhenCellUpdates(t *testing.T) {
	table := MakeTable("test", nil)
	table.EditTableValue(1, 1, "5")
	table.EditTableValue(2, 2, "5")
	sumCell := table.EditTableValue(5, 0, "=sum(A1:C3)")
	ch := MakeValueChannel()
	sumCell.Subscribe(ch)
	table.EditTableValue(1, 1, "7")
	<- ch
	if sumCell.DisplayValue != "12" {
		t.Error("Summation cell not updated correctly.")
	}
}

func TestSumUpdatesWhenEmptyCellUpdates(t *testing.T) {
	table := MakeTable("test", nil)
	table.EditTableValue(1, 1, "5")
	sumCell := table.EditTableValue(5, 0, "=sum(A1:C3)")
	ch := MakeValueChannel()
	sumCell.Subscribe(ch)
	table.EditTableValue(2, 2, "5")
	<- ch
	if sumCell.DisplayValue != "10" {
		t.Error("Sum did not update with empty cell correctly.")
	}
}

func TestSumUpdatesCascadeAcrossSums(t *testing.T) {
	table := MakeTable("test", nil)
	table.EditTableValue(1, 4, "5")
	table.EditTableValue(1, 1, "=sum(E1:F2)")
	table.EditTableValue(2, 1, "5")
	sumCell := table.EditTableValue(5, 0, "=sum(A1:C3)")
	ch := MakeValueChannel()
	sumCell.Subscribe(ch)
	if sumCell.DisplayValue != "10" {
		t.Error("Initial summation not calculated correctly.")
	}
	table.EditTableValue(1, 4, "10")
	<- ch // final value of 15
	if sumCell.DisplayValue != "15" {
		t.Error("Final summation not calculated correctly.")
	}
}

func TestTableNotifiesSubscribersWhenCellsChange(t *testing.T) {
	ch := MakeValueChannel()
	table := MakeTable("test", nil)
	table.SubscribeToTable(ch)
	cell := table.EditTableValue(1, 1, "1")
	message := <- ch
	if cell != message.cell {
		t.Error("Table not correctly notifying listeners.")
	}
}

func TestFormulaRangeChange(t *testing.T) {
	ch := MakeValueChannel()
	sumCh := MakeValueChannel()
	table := MakeTable("test", nil)
	cell := table.EditTableValue(1, 1, "5")
	cell.Subscribe(ch)
	sumCell := table.EditTableValue(5, 0, "=sum(A1:C3)")
	sumCell.Subscribe(sumCh)
	if sumCell.DisplayValue != "5" {
		t.Error("Sum cell not calculated correctly.")
	}
	table.EditTableValue(5, 0, "=sum(E1:G3)")
	<- sumCh
	if sumCell.DisplayValue != "0" {
		t.Error("Sum cell not calclated correctly after sum range update.")
	}
	table.EditTableValue(1, 1, "5")
	<- ch
	if sumCell.DisplayValue != "0" {
		t.Error("Sum cell recalculated using stale range.")
	}
	table.EditTableValue(1, 5, "5")
	<- sumCh
	if sumCell.DisplayValue != "5" {
		t.Error("Formula did not update correctly with new range.")
	}
}

/*
This will deadlock if notifications from table aren't sent correctly
 */
func TestTableNotifiesWhenCellsUpdated(t *testing.T) {
	ch := MakeValueChannel()
	table := MakeTable("test", nil)
	table.SubscribeToTable(ch)
	table.EditTableValue(1, 1, "1")
	table.EditTableValue(2, 1, "1")
	table.EditTableValue(5, 0, "=sum(A1:C3)")
	<- ch
	<- ch
	<- ch
}

func TestCascadingUpdatesSendsNotificationsThroughTable(t *testing.T) {
	ch := MakeValueChannel()
	table := MakeTable("test", nil)
	table.SubscribeToTable(ch)
	table.EditTableValue(1, 5, "5")
	table.EditTableValue(1, 1, "=sum(E1:G3)")
	table.EditTableValue(2, 1, "5")
	sumCell := table.EditTableValue(5, 0, "=sum(A1:C3)")
	<- ch
	<- ch
	<- ch
	message := <- ch
	if message.cell != sumCell {
		t.Error("Table not correctly handing subscriptions.")
	}
	table.EditTableValue(1, 5, "10")
	<- ch // constant updated
	<- ch // first sum updated
	message = <- ch // second sum updated
	if message.cell != sumCell {
		t.Error("Cascading formulas not notifying table.")
	}
}
