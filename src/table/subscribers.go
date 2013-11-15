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
)

/*
This naming convention should obviously be cleaned up, but observers
are looking for changes to an item while subscribers are subscribed to a
specific event on any channel
 */
type observers struct {
	observers		[]node.IMessage
}

func (o *observers) isObserversInList(cmd node.IMessage) bool {
	for _, c := range o.observers {
		if c.Equal(cmd) {
			return true
		}
	}
	return false
}

func (o *observers) removeObserver(cmd node.IMessage) {
	i := -1
	for index, command := range o.observers {
		if command == cmd {
			i = index
			break
		}
	}
	if i == -1 {
		return
	}

	o.observers = append(o.observers[:i], o.observers[i+1:]...)
}

func (o *observers) notifyObservers(operation string, ch chan node.IMessage, bytes []byte) {
	for _, cmd := range o.observers {
		go func () {
			response := node.MakeResponse(cmd, bytes)
			ch <- response
		}()
	}
}

func (o *observers) addObserver(cmd node.IMessage) {
	if (o.isObserversInList(cmd)) {
		return
	}
	o.observers = append(o.observers, cmd)
}

func MakeObservers() *observers {
	obs := new(observers)
	obs.observers = make([]node.IMessage, 0)
	return obs
}
