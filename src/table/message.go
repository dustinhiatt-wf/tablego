/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/4/13
 * Time: 10:57 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"time"
)

type valuemessage struct {
	operation 	string
	cell		*cell
	table		*table
	tableRange	*tablerange
	valueRange	*valuerange
	messageId	string
	timestamp	int
}

func MakeValueMessage(operation, messageId string, cell *cell, vr *valuerange, tr *tablerange, table *table) *valuemessage {
	vm := new(valuemessage)
	vm.operation = operation
	vm.messageId = messageId
	vm.cell = cell
	vm.table = table
	vm.tableRange = tr
	vm.valueRange = vr
	vm.timestamp = time.Now().Nanosecond()
	return vm
}

func (vm *valuemessage) Copy() *valuemessage {
	copy := new(valuemessage)
	*copy = *vm
	return copy
}

type tablemessage struct {
	operation	string
}

func MakeValueChannel() chan *valuemessage {
	return make(chan *valuemessage)
}
