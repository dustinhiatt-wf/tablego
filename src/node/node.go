package node

import (
//	"log"
)

const (
	ChildInitialized	= "childInitialized"
)

type ICommunicationHandler interface {
	OnMessageFromParent(msg IMessage)
	OnMessageFromChild(msg IMessage)
}

type INodeFactory interface {
	GetChild(coords ICoordinates) 											IChild
	MakeChildNode(parentChannel IChild, childCoordinates ICoordinates) 		INode
	DumpInformation(c ICoordinates)											string
}

type IChild interface {
	Channel()											IChannel
	IsInitialized()										bool
	PendingRequests()									map[string]isubscribers
	Subscribe(operation string, ch chan IMessage)
	SendNotification(operation string, message IMessage, clear bool)
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
func (c *child) Subscribe(operation string, ch chan IMessage) {
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

func (c *child) SendNotification(operation string, message IMessage, clear bool) {
	subs, ok := c.childPendingRequests[operation]
	if ok {
		subs.notifySubscribers(message, clear)
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

func MakeIChild() IChild {
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
	Send(ch chan IMessage, msg IMessage)
	listenToParent(parentChannel IChannel, startSignal chan IMessage)
	Coordinates()													ICoordinates
	ParentCoordinates()												ICoordinates
	Initialize()
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
	} else if n.ParentCoordinates() == nil && msg.TargetCoordinates() != nil {
		return false
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
				go n.Send(n.Parent().ChildToParent(), message)
				continue
			} else if !n.isMessageIntendedForMe(message) {
				child := n.nodeFactory.GetChild(message.TargetCoordinates())
				if child != nil {
					go n.Send(child.Channel().ParentToChild(), message)
				}
				continue
			}

			if message.Operation() == ChildInitialized {
				child.setInitialized(true)
				go child.SendNotification(ChildInitialized, message, true)
			} else {
				go n.communicationHandler.OnMessageFromChild(message)
			}
		}
	}
}

func (n *Node) listenToParent(parentChannel IChannel, startSignal chan IMessage) {
	startSignal <- nil
	if parentChannel == nil {
		return // we are the top level parent
	}

	for {
		select {
		case message := <- parentChannel.ParentToChild():
			if !n.isMessageIntendedForMe(message) {
				go func () {
					ch := MakeMessageChannel()
					n.GetOrCreateChild(ch, message.TargetCoordinates())
					<- ch // child initialized
					child := n.nodeFactory.GetChild(message.TargetCoordinates())
					child.Channel().ParentToChild() <- message
				}()
				continue
			}

			go n.communicationHandler.OnMessageFromParent(message)
		}
	}
}

func (n *Node) Initialize() {
	if n.parent != nil {
		go n.Send(n.parent.ChildToParent(), MakeCommand(ChildInitialized, n.ParentCoordinates(), n.Coordinates(), nil))
	}
}

func (n *Node) Send(ch chan IMessage, msg IMessage) {
	ch <- msg
}

func (n *Node) isOwnCoordinates(other ICoordinates) bool {
	return n.coordinates.Equal(other)
}

func (n *Node) isMessageIntendedForMe(msg IMessage) bool {
	return n.Coordinates().Equal(msg.TargetCoordinates())
}

func (n *Node) createChild(observer chan IMessage, coords ICoordinates) {
	child := MakeIChild()
	child.Subscribe(ChildInitialized, observer)
	n.nodeFactory.MakeChildNode(child, coords)
	go n.listenToChild(child)
}

func (n *Node) GetOrCreateChild(observer chan IMessage, coords ICoordinates) {
	child := n.nodeFactory.GetChild(coords)
	if child == nil {
		n.createChild(observer, coords)
	} else if !child.IsInitialized() { //currently being loaded
		child.Subscribe(ChildInitialized, observer)
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
	ch := MakeMessageChannel()
	go node.listenToParent(node.parent, ch)
	<- ch
	return node
}
