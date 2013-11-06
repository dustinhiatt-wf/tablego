/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/4/13
 * Time: 9:31 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"strconv"
	"time"
	"strings"
//	"log"
)

const letters string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

type cell struct {
	row					int
	column				int
	value				string
	valuechannel		chan *valuemessage
	timestamp			int
	table				*table
	isFormula			bool
	DisplayValue		string
	subscribers			*subscribers
	observers			[]chan *valuemessage
}

func (c *cell) AsInt() (int64, bool) {
	i, err := strconv.ParseInt(c.DisplayValue, 0, 64) // 64-bit, base implied from string
	if err != nil {
		return 0, false
	}
	return i, true
}


func (c *cell) AsFloat() (float64, bool) {
	f, err := strconv.ParseFloat(c.DisplayValue, 64) //64-bit parse
	if err != nil {
		return 0, false
	}
	return f, true
}


func (c *cell) IsInt() bool {
	_, success := c.AsInt()
	return success
}


func (c *cell) IsFloat() bool {
	_, success := c.AsFloat()
	return success
}

func (c *cell) SetValue(value string) {
	if c.value == value {
		return
	}
	c.value = value
	c.DisplayValue = value
	if c.table != nil && c.table.isinitialized {
		initialize(c)
	}
	if !c.isFormula {
		go c.subscribers.notifySubscribers(&valuemessage{"updated", c})
	}
}

func (c *cell) Subscribe(ch chan *valuemessage) {
	c.subscribers.append(ch)
}

func initialize(c *cell) {
	if c.value == "" {
		return
	}
	parseValueForFormula(c)
}

func listen(c *cell, ch <- chan *valuemessage) {
	for {
		select {
		case message := <- ch:
			switch message.operation {
			case "updated":
				go calculate(c)
			case "unsubscribe":
				return
			}
		}
	}
}

func subscribeToRange(c *cell, cr *cellrange) {
	unsubscribeObservers(c)
	for i := cr.startRow; i < cr.stopRow; i++ {
		for j := cr.startColumn; j < cr.stopColumn; j++ {
			ch := MakeValueChannel()
			c.table.Subscribe(i, j, ch)
			c.observers = append(c.observers, ch)
			go listen(c, ch)
		}
	}
}

func unsubscribeObservers(c *cell) {
	for _, ch := range c.observers {
		ch <- &valuemessage{"unsubscribe", nil}
		close(ch)
	}
	c.observers = make([]chan *valuemessage, 0)
}

func calculate(c *cell) *cellrange {
	funcParts := parseFormula(c.value)
	var cr *cellrange
	value := ""
	switch funcParts[0] {
	case "sum":
		cr, value = sum(c, funcParts[1])
	}
	if value != c.DisplayValue {
		c.DisplayValue = value
		go c.subscribers.notifySubscribers(&valuemessage{"updated", c})
	}
	return cr
}

func listenToTable(c *cell) {
	for {
		select {
		case message := <- c.table.tablechannel:
			switch message.operation {
			case "initialized":
				go initialize(c)
			}
		}
	}
}

func parseValueForFormula(c *cell) {
	if strings.HasPrefix(c.value, "=") {
		c.isFormula = true
		cr := calculate(c)
		if cr != nil {
			subscribeToRange(c, cr)
		}
	} else {
		c.DisplayValue = c.value
	}
}

func MakeCell(row int, column int, value string, table *table, listener chan *valuemessage) *cell {
	cell := new(cell)
	cell.valuechannel = MakeValueChannel()
	cell.row = row
	cell.column = column
	cell.timestamp = time.Now().Nanosecond()
	cell.table = table
	cell.subscribers = MakeSubscribers()
	if listener != nil {
		cell.subscribers.append(listener)
	}
	cell.observers = make([]chan *valuemessage, 0)
	cell.SetValue(value)
	if table != nil {
		go listenToTable(cell)
		if table.isinitialized {
			initialize(cell)
		}
	}
	return cell
}

