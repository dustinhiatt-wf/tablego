/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/5/13
 * Time: 1:59 PM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"testing"
)

func TestIsSubscriberInSubscribers(t *testing.T) {
	s := MakeSubscribers()
	ch := MakeValueChannel()
	s.subscribers = append(s.subscribers, ch)
	if !s.isSubscriberInList(ch) {
		t.Error("Subscriber is actually in list.")
	}

	if s.isSubscriberInList(MakeValueChannel()) {
		t.Error("Subscriber is not in list.")
	}
}

func TestAppendSubscriber(t *testing.T) {
	s := MakeSubscribers()
	ch := MakeValueChannel()
	s.append(ch)
	if len(s.subscribers) != 1 {
		t.Error("Channel not appended properly")
	}
	s.append(ch)
	if len(s.subscribers) != 1 {
		t.Error("Identical channels added.")
	}
}

func TestRemoveSubscriber(t *testing.T) {
	s := MakeSubscribers()
	ch := MakeValueChannel()
	s.append(ch)
	s.remove(ch)
	if len(s.subscribers) != 0 {
		t.Error("Channel not removed correctly.")
	}
}

/*
This would deadlock if a message was sent
 */
func TestNotifyEmptySubscribers(t *testing.T) {
	s := MakeSubscribers()
	s.notifySubscribers(MakeValueMessage(Updated, "", new(cell), nil, nil, nil), false)
}

/*
This would deadlock if a message was not sent
 */
func TestNotifySubscriber(t *testing.T) {
	s := MakeSubscribers()
	ch := MakeValueChannel()
	s.append(ch)
	go s.notifySubscribers(MakeValueMessage(Updated, "", new(cell), nil, nil, nil), false)
	<- ch
}

func TestNotifyClosedChannel(t *testing.T) {
	s := MakeSubscribers()
	ch := MakeValueChannel()
	chTwo := MakeValueChannel()
	s.append(ch)
	s.append(chTwo)
	close(ch)
	s.notifySubscribers(MakeValueMessage(Updated, "", new(cell), nil, nil, nil), false)
	<- chTwo
	if len(s.subscribers) != 1 {
		t.Error("Exception did not remove subscriber.")
	}
}
