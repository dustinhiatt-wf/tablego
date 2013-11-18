package table

import (
	"node"
	"encoding/json"
)

type subscribePayload struct {
	ISerializable
	Row			int
	Column		int
	TableId		string
	HasCellLoc	bool //cheating because of json conversion of negative numbers
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
	sp.Row = row
	sp.Column = column
	sp.TableId = tableId
	sp.HasCellLoc = hasCellLoc
	return sp
}

func makeSubscribePayloadFromBytes(bytes []byte) *subscribePayload {
	var sp subscribePayload
	json.Unmarshal(bytes, &sp)
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
		if c.Column == sp.Column && c.Row == sp.Row && c.TableId == sp.TableId {
			return true
		}
	}
	return false
}

func (o *observers) removeObserver(sp *subscribePayload) {
	i := -1
	for index, c := range o.observers {
		if c.Column == sp.Column && c.Row == sp.Row && c.TableId == sp.TableId {
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
	for _, observer := range o.observers {
		go func(obs *subscribePayload) {
			var destination node.ICoordinates
			if !obs.HasCellLoc {
				destination = MakeCoordinates(obs.TableId, nil)
			} else {
				destination = MakeCoordinates(obs.TableId, MakeCellLocation(obs.Row, obs.Column))
			}
			cmd := node.MakeCommand(operation, destination, sourceCoordinates, bytes)
			ch <- cmd
		}(observer)
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
