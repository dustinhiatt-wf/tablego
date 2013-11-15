/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/5/13
 * Time: 1:56 PM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"node"
	"encoding/json"
)

type subscribePayload struct {
	ISerializable
	row			int
	column		int
	tableId		string
	hasCellLoc	bool //cheating because of json conversion of negative numbers
}

func (sp *subscribePayload) ToBytes() []byte {
	res, err := json.Marshal(sp)
	if err != nil {
		return nil
	}
	return res
}

func makeSubscribePayload(tableId string, row, column int, hasCellLoc bool) *subscribePayload {
	sp := new(subscribePayload)
	sp.row = row
	sp.column = column
	sp.tableId = tableId
	sp.hasCellLoc = hasCellLoc
	return sp
}

func makeSubscribePayloadFromBytes(bytes []byte) *subscribePayload {
	var sp subscribePayload
	json.Unmarshal(bytes, sp)
	return &sp
}

/*
This naming convention should obviously be cleaned up, but observers
are looking for changes to an item while subscribers are subscribed to a
specific event on any channel
*/
type observers struct {
	observers []*subscribePayload
}

func (o *observers) isObserversInList(sp *subscribePayload) bool {
	for _, c := range o.observers {
		if c.column == sp.column && c.row == sp.row && c.tableId == sp.tableId {
			return true
		}
	}
	return false
}

func (o *observers) removeObserver(sp *subscribePayload) {
	i := -1
	for index, c := range o.observers {
		if c.column == sp.column && c.row == sp.row && c.tableId == sp.tableId {
			i = index
			break
		}
	}
	if i == -1 {
		return
	}

	o.observers = append(o.observers[:i], o.observers[i+1:]...)
}

func (o *observers) notifyObservers(operation string, ch chan node.IMessage, bytes []byte, sourceCoordinates node.ICoordinates) {
	for _, cmd := range o.observers {
		go func() {
			var destination node.ICoordinates
			if !cmd.hasCellLoc {
				destination = MakeCoordinates(cmd.tableId, nil)
			} else {
				destination = MakeCoordinates(cmd.tableId, MakeCellLocation(cmd.row, cmd.column))
			}
			cmd := node.MakeCommand(operation, destination, sourceCoordinates, bytes)
			ch <- cmd
		}()
	}
}

func (o *observers) addObserver(sp *subscribePayload) {
	if o.isObserversInList(sp) {
		return
	}
	o.observers = append(o.observers, sp)
}

func MakeObservers() *observers {
	obs := new(observers)
	obs.observers = make([]*subscribePayload, 0)
	return obs
}
