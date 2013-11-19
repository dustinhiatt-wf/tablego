package node

type isubscribers interface {
	Subscribers()							[]chan IMessage
	isSubscriberInList(ch chan IMessage) 	bool
	clear()
	remove(ch chan IMessage)
	append(ch chan IMessage)
	notifySubscribers(message IMessage, clear bool)
}

type subscribers struct {
	isubscribers
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

func (s *subscribers) Subscribers() []chan IMessage {
	return s.subscribers
}

func MakeSubscribers() isubscribers {
	s := new(subscribers)
	s.subscribers = make([]chan IMessage, 0)
	return s
}

