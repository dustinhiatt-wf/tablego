/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/5/13
 * Time: 1:56 PM
 * To change this template use File | Settings | File Templates.
 */
package table

type subscribers struct {
	subscribers		[]chan IMessage
}

func (s *subscribers) isSubscriberInList(ch chan IMessage) bool {
	for _, sub := range s.subscribers {
		if sub == ch {
			return true
		}
	}
	return false
}

func (s *subscribers) clear() {
	s.subscribers = make([]chan IMessage, 0)
}

func (s *subscribers) remove(ch chan IMessage) {
	if !s.isSubscriberInList(ch) {
		return
	}
	i := -1
	for index, channel := range s.subscribers {
		if channel == ch {
			i = index
			break
		}
	}
	s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
}

func (s *subscribers) append(ch chan IMessage) {
	if !s.isSubscriberInList(ch) {
		s.subscribers = append(s.subscribers, ch)
	}
}

func internalNotify(s *subscribers, ch chan IMessage, message IMessage) {
	defer func() {
		if err := recover(); err != nil {
			s.remove(ch)
		}
	}()
	ch <- message
}

func (s *subscribers) notifySubscribers(message IMessage, clear bool) {
	if len(s.subscribers) == 0 {
		return
	}

	for _, ch := range s.subscribers {
		go internalNotify(s, ch, message)
	}

	if clear {
		s.clear()
	}
}

func MakeSubscribers() *subscribers {
	s := new(subscribers)
	s.subscribers = make([]chan IMessage, 0)
	return s
}

/*
This naming convention should obviously be cleaned up, but observers
are looking for changes to an item while subscribers are subscribed to a
specific event on any channel
 */
type observers struct {
	observers		[]ICommand
}

func (o *observers) isObserversInList(cmd ICommand) bool {
	for _, c := range o.observers {
		if c.Equal(cmd) {
			return true
		}
	}
	return false
}

func (o *observers) removeObserver(cmd ICommand) {
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

func (o *observers) notifyObservers(operation string, ch chan IMessage, bytes []byte) {
	for _, cmd := range o.observers {
		go func () {
			response := MakeResponse(cmd, bytes)
			response.operation = operation
			ch <- response
		}()
	}
}

func (o *observers) addObserver(cmd ICommand) {
	if (o.isObserversInList(cmd)) {
		return
	}
	o.observers = append(o.observers, cmd)
}

func MakeObservers() *observers {
	obs := new(observers)
	obs.observers = make([]ICommand, 0)
	return obs
}
