/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/4/13
 * Time: 9:17 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	//"log"
)

type table struct {
	id				string
	cells			map[int]map[int]*cell
	rows			map[int]chan *valuemessage
	columns			map[int]chan *valuemessage
	isinitialized	bool
	tablechannel	chan *tablemessage
	subscribers		*subscribers
	orchestrator	*orchestrator
}

/*
Listens for cell changes and notifies anyone listening to the table
 */
func listenToCells(t *table, ch chan *valuemessage) {
	for {
		select {
		case message := <- ch:
			t.subscribers.notifySubscribers(message)
		}
	}
}

func getCell(table *table, row, column int) *cell {
	_, ok := table.cells[row]
	if ok {
		cell, ok := table.cells[row][column]
		if ok {
			return cell
		}
	}
	return nil
}

func tableHasCell(table *table, row, column int) bool {
	return getCell(table, row, column) != nil
}

func MakeTable(id string, orchestrator *orchestrator) *table {
	table := new(table)
	table.id = id
	cells := make(map[int]map[int]*cell)
	table.cells = cells
	table.tablechannel = MakeTableChannel()
	table.subscribers = MakeSubscribers()
	table.orchestrator = orchestrator
	table.isinitialized = true
	return table
}

func (t *table) EditTableValue(row, column int, value string) *cell {
	row_item, ok := t.cells[row]
	if ok {
		cell, ok := row_item[column]
		if ok {
			cell.SetValue(value)
		} else {
			ch := MakeValueChannel()
			cell := MakeCell(row, column, value, t, ch)
			go listenToCells(t, ch)
			t.cells[row][column] = cell
		}
	} else {
		t.cells[row] = make(map[int]*cell)
		ch := MakeValueChannel()
		go listenToCells(t, ch)
		cell := MakeCell(row, column, value, t, ch)
		t.cells[row][column] = cell
	}
	return t.cells[row][column]
}

func (t *table) Subscribe(row, column int, ch chan *valuemessage) {
	cell := getCell(t, row, column)
	if cell != nil {
		cell.Subscribe(ch)
		return
	}
	cell = t.EditTableValue(row, column, "")
	cell.Subscribe(ch)
}

func (t *table) SubscribeToTable(ch chan *valuemessage) {
	t.subscribers.append(ch)
}

func (t *table) GetValueAt(row, column int) *cell {
	return getCell(t, row, column)
}

func (t *table) GetRangeByRowAndColumn(startRow, stopRow, startColumn, stopColumn int) *tablerange {
	cr := &cellrange{startRow, stopRow + 1, startColumn, stopColumn + 1, t.id}
	return t.GetRangeByCellRange(cr)
}

func (t *table) GetRangeByCellRange(cr *cellrange) *tablerange {
	return MakeTableRange(t.cells, cr)
}
