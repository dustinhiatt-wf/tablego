/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/5/13
 * Time: 1:56 PM
 * To change this template use File | Settings | File Templates.
 */
package table

type subscribers struct {
	subscribers		[]chan *valuemessage
}

func (s *subscribers) isSubscriberInList(ch chan *valuemessage) bool {
	for _, sub := range s.subscribers {
		if sub == ch {
			return true
		}
	}
	return false
}

func (s *subscribers) clear() {
	s.subscribers = make([]chan *valuemessage, 0)
}

func (s *subscribers) remove(ch chan *valuemessage) {
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

func (s *subscribers) append(ch chan *valuemessage) {
	if !s.isSubscriberInList(ch) {
		s.subscribers = append(s.subscribers, ch)
	}
}

func internalNotify(s *subscribers, ch chan *valuemessage, message *valuemessage) {
	defer func() {
		if err := recover(); err != nil {
			s.remove(ch)
		}
	}()
	ch <- message
}

func (s *subscribers) notifySubscribers(message *valuemessage, clear bool) {
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
	s.subscribers = make([]chan *valuemessage, 0)
	return s
}
