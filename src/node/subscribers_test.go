/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/14/13
 * Time: 10:41 AM
 * To change this template use File | Settings | File Templates.
 */
package node

import (
	"testing"
)

func TestIsSubscriberInSubscribers(t *testing.T) {
	s := MakeSubscribers()
	ch := MakeMessageChannel()
	s.append(ch)
	if !s.isSubscriberInList(ch) {
		t.Error("Subscriber is actually in list.")
	}

	if s.isSubscriberInList(MakeMessageChannel()) {
		t.Error("Subscriber is not in list.")
	}
}

func TestAppendSubscriber(t *testing.T) {
	s := MakeSubscribers()
	ch := MakeMessageChannel()
	s.append(ch)
	if len(s.Subscribers()) != 1 {
		t.Error("Channel not appended properly")
	}
	s.append(ch)
	if len(s.Subscribers()) != 1 {
		t.Error("Identical channels added.")
	}
}

func TestRemoveSubscriber(t *testing.T) {
	s := MakeSubscribers()
	ch := MakeMessageChannel()
	s.append(ch)
	s.remove(ch)
	if len(s.Subscribers()) != 0 {
		t.Error("Channel not removed correctly.")
	}
}

/*
This would deadlock if a message was sent
 */
func TestNotifyEmptySubscribers(t *testing.T) {
	s := MakeSubscribers()
	s.notifySubscribers(makeMessage("test", nil, nil, nil), false)
}

/*
This would deadlock if a message was not sent
 */
func TestNotifySubscriber(t *testing.T) {
	s := MakeSubscribers()
	ch := MakeMessageChannel()
	s.append(ch)
	go s.notifySubscribers(makeMessage("test", nil, nil, nil), false)
	<- ch
}

func TestNotifyClosedChannel(t *testing.T) {
	s := MakeSubscribers()
	ch := MakeMessageChannel()
	chTwo := MakeMessageChannel()
	s.append(ch)
	s.append(chTwo)
	close(ch)
	s.notifySubscribers(makeMessage("test", nil, nil, nil), false)
	<- chTwo
	if len(s.Subscribers()) != 1 {
		t.Error("Exception did not remove subscriber.")
	}
}

