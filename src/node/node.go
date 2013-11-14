package node

import (
//	"log"
)

const (
	ChildInitialized	= "childInitialized"
)

type ICommunicationHandler interface {
	onMessageFromParent(msg IMessage)
	onMessageFromChild(msg IMessage)
}

type INodeFactory interface {
	GetChild(coords ICoordinates) 											IChild
	makeChildNode(parentChannel IChild, childCoordinates ICoordinates) 	INode
}

type IChild interface {
	Channel()											IChannel
	IsInitialized()										bool
	PendingRequests()									map[string]isubscribers
	subscribe(operation string, ch chan IMessage)
	sendNotification(operation string, message IMessage)
	setInitialized(value bool)
}

type child struct {
	IChild
	channel						IChannel
	isInitialized				bool
	childPendingRequests		map[string]isubscribers
}

/*
If ch is nil, the function returns
 */
func (c *child) subscribe(operation string, ch chan IMessage) {
	if ch == nil {
		return
	}

	subs, ok := c.childPendingRequests[operation]
	if !ok {
		c.childPendingRequests[operation] = MakeSubscribers()
		subs =c.childPendingRequests[operation]
	}
	subs.append(ch)
}

func (c *child) sendNotification(operation string, message IMessage) {
	subs, ok := c.childPendingRequests[operation]
	if ok {
		subs.notifySubscribers(message, true)
	}
}

func (c *child) setInitialized(value bool) {
	c.isInitialized = value
}

func (c *child) Channel() IChannel {
	return c.channel
}

func (c *child) IsInitialized() bool {
	return c.isInitialized
}

func makeIChild() IChild {
	child := new(child)
	child.channel = makeIChannel()
	child.isInitialized = false // not necessary, but verbose
	child.childPendingRequests = make(map[string]isubscribers)
	return child
}

type IChannel interface {
	ParentToChild() chan IMessage
	ChildToParent() chan IMessage
}

type channel struct {
	IChannel
	parentToChild	chan IMessage
	childToParent	chan IMessage
}

func (c *channel) ParentToChild() chan IMessage {
	return c.parentToChild
}

func (c *channel) ChildToParent() chan IMessage {
	return c.childToParent
}

func makeIChannel() IChannel {
	ch := new(channel)
	ch.parentToChild = make(chan IMessage)
	ch.childToParent = make(chan IMessage)
	return ch
}

type INode interface {
	GetOrCreateChild(observer chan IMessage, coords ICoordinates)
	Parent()														IChannel
	createChild(observer chan IMessage, coords ICoordinates)
	listenToChild(child IChild)
	isOwnCoordinates(coords ICoordinates)							bool
	isMessageIntendedForMe(msg IMessage)							bool
	isMessageIntendedForParent(msg IMessage)						bool
	send(ch chan IMessage, msg IMessage)
	listenToParent(parentChannel IChannel)
	Coordinates()													ICoordinates
	ParentCoordinates()												ICoordinates
	initialize()
}

type Node struct {
	INode
	parent					IChannel
	coordinates				ICoordinates
	parentCoordinates		ICoordinates
	communicationHandler	ICommunicationHandler
	nodeFactory				INodeFactory
}

func (n *Node) Coordinates() ICoordinates {
	return n.coordinates
}

func (n *Node) ParentCoordinates() ICoordinates {
	return n.parentCoordinates
}

func (n *Node) Parent() IChannel {
	return n.parent
}

func (n *Node) isMessageIntendedForParent(msg IMessage) bool {
	if n.ParentCoordinates() == nil && msg.TargetCoordinates() == nil {
		return true
	}
	if n.ParentCoordinates().Equal(msg.TargetCoordinates()) {
		return true
	}
	return false
}

func (n *Node) listenToChild(child IChild) {
	for {
		select {
		case message := <- child.Channel().ChildToParent():
			if n.isMessageIntendedForParent(message) {
				go n.send(n.Parent().ChildToParent(), message)
				continue
			} else if !n.isMessageIntendedForMe(message) {
				child := n.nodeFactory.GetChild(message.TargetCoordinates())
				if child != nil {
					go n.send(child.Channel().ParentToChild(), message)
				}
				continue
			}

			if message.Operation() == ChildInitialized {
				child.setInitialized(true)
				go child.sendNotification(ChildInitialized, message)
			} else {
				go n.communicationHandler.onMessageFromChild(message)
			}
		}
	}
}

func (n *Node) listenToParent(parentChannel IChannel) {
	if parentChannel == nil {
		return // we are the top level parent
	}

	for {
		select {
		case message := <- parentChannel.ParentToChild():
			if !n.isMessageIntendedForMe(message) {
				child := n.nodeFactory.GetChild(message.TargetCoordinates())
				if child != nil {
					go n.send(child.Channel().ParentToChild(), message)
				}
				continue
			}

			go n.communicationHandler.onMessageFromParent(message)
		}
	}
}

func (n *Node) initialize() {
	go n.listenToParent(n.parent)
	if n.parent != nil {
		go n.send(n.parent.ChildToParent(), MakeCommand(ChildInitialized, n.ParentCoordinates(), n.Coordinates(), nil))
	}
}

func (n *Node) send(ch chan IMessage, msg IMessage) {
	ch <- msg
}

func (n *Node) isOwnCoordinates(other ICoordinates) bool {
	return n.coordinates.Equal(other)
}

func (n *Node) isMessageIntendedForMe(msg IMessage) bool {
	return n.Coordinates().Equal(msg.TargetCoordinates())
}

func (n *Node) createChild(observer chan IMessage, coords ICoordinates) {
	child := makeIChild()
	child.subscribe(ChildInitialized, observer)
	go n.listenToChild(child)
	go n.nodeFactory.makeChildNode(child, coords)
}

func (n *Node) GetOrCreateChild(observer chan IMessage, coords ICoordinates) {
	child := n.nodeFactory.GetChild(coords)
	if child == nil {
		go n.createChild(observer, coords)
	} else if !child.IsInitialized() { //currently being loaded
		child.subscribe(ChildInitialized, observer)
	} else { // table is loaded and ready
		go func () {
			observer <- makeMessage(ChildInitialized, n.ParentCoordinates(), n.Coordinates(), nil)
		}()
	}
}

func MakeNode(parentChannel IChannel, coordinates, parentCoordinates ICoordinates, communicationHandler ICommunicationHandler, nodeFactory INodeFactory) INode {
	node := new(Node)
	node.parent = parentChannel
	node.coordinates = coordinates
	node.parentCoordinates = parentCoordinates
	node.communicationHandler = communicationHandler
	node.nodeFactory = nodeFactory
	node.initialize()
	return node
}
