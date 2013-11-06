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
	id						string
	cells					map[int]map[int]*cell
	rows					map[int]chan *valuemessage
	columns					map[int]chan *valuemessage
	isinitialized			bool
	tablechannel			chan *valuemessage
	subscribers				*subscribers
	orchestrator			*orchestrator
	orchestratorChannel 	chan *valuemessage
}

/*
Listens for cell changes and notifies anyone listening to the table
 */
func listenToCells(t *table, ch chan *valuemessage) {
	for {
		select {
		case message := <- ch:
			t.subscribers.notifySubscribers(message, false)
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

func listenToOrchestrator(t *table) {
	for {
		select {
		case message := <- t.orchestratorChannel:
			switch message.operation {
			case "getTable":
				t.orchestratorChannel <- MakeValueMessage(GetTable, "", nil, nil, nil, t)
			}
		}
	}
}

func MakeTable(id string, orchestrator *orchestrator, ch chan *valuemessage) *table {
	table := new(table)
	table.id = id
	cells := make(map[int]map[int]*cell)
	table.cells = cells
	table.tablechannel = MakeValueChannel()
	table.subscribers = MakeSubscribers()
	table.orchestrator = orchestrator
	table.isinitialized = true
	table.orchestratorChannel = ch
	go func () {
		ch <- MakeValueMessage("tableOpened", "", nil, nil, nil, table)
	}()
	go listenToOrchestrator(table)
	return table
}

func (t *table) EditTableValue(row, column int, value string, ch chan *valuemessage) {
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
	go func () {
		ch <- MakeValueMessage(EditCell, "", t.cells[row][column], nil, nil, nil)
	}()
}

func (t *table) Subscribe(ch chan *valuemessage) {
	t.subscribers.append(ch)
}

func (t *table) GetValueAt(row, column int, ch chan *valuemessage) {
	cell := getCell(t, row, column)
	go func () {
		ch <- MakeValueMessage(GetValueAt, "", cell, nil, nil, nil)
	}()
}

func (t *table) GetRangeByRowAndColumn(startRow, stopRow, startColumn, stopColumn int, ch chan *valuemessage) {
	cr := &cellrange{startRow, stopRow + 1, startColumn, stopColumn + 1, t.id}
	t.GetRangeByCellRange(cr, ch)
}

func (t *table) GetRangeByCellRange(cr *cellrange, ch chan *valuemessage) {
	tr := MakeTableRange(t.cells, cr)
	go func () {
		ch <- MakeValueMessage(GetCellRange, "", nil, nil, tr, nil)
	}()
}
