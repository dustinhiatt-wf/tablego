/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/4/13
 * Time: 10:57 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"node"
)

type ISerializable interface {
	ToBytes() []byte
}

type ITableCoordinates interface {
	node.ICoordinates
	TableId() string
	CellLocation() ICellLocation
}

type coordinates struct {
	ITableCoordinates
	tableId      string
	cellLocation ICellLocation
}

func (c *coordinates) Equal(other node.ICoordinates) bool {
	o := other.(ITableCoordinates)
	if c.cellLocation == nil && o.CellLocation() != nil {
		return false
	} else if c.cellLocation == nil && o.CellLocation() == nil {
		return c.tableId == o.TableId()
	}
	return c.tableId == o.TableId() && c.cellLocation.Equal(o.CellLocation())
}

func (c *coordinates) CellLocation() ICellLocation {
	return c.cellLocation
}

func (c *coordinates) TableId() string {
	return c.tableId
}

type ICellLocation interface {
	Row() int
	Column() int
	Equal(other ICellLocation) bool
}

type cellLocation struct {
	cellRow    int
	cellColumn int
}

func (cl *cellLocation) Equal(other ICellLocation) bool {
	if other == nil {
		return false
	}
	return cl.cellRow == other.Row() && cl.cellColumn == other.Column()
}

func (cl *cellLocation) Row() int {
	return cl.cellRow
}

func (cl *cellLocation) Column() int {
	return cl.cellColumn
}

func MakeCellLocation(row, column int) ICellLocation {
	cl := &cellLocation{row, column}
	return cl
}

func MakeCoordinates(tableId string, cellLocation ICellLocation) node.ICoordinates {
	c := new(coordinates)
	c.tableId = tableId
	c.cellLocation = cellLocation
	return c
}

type addToChildMessage struct {
	row           int
	column        int
	tableId       string
	child         node.IChild
	returnChannel chan node.IChild
}

func makeAddToChildMessage(row, column int, tableId string, child node.IChild) *addToChildMessage {
	atcm := new(addToChildMessage)
	atcm.row = row
	atcm.column = column
	atcm.child = child
	atcm.tableId = tableId
	atcm.returnChannel = make(chan node.IChild)
	return atcm
}
