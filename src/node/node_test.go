package node

import (
	"testing"
	"reflect"
)

type ILocationCoordinates interface {
	ICoordinates
	Location()		string
}

type coordinate struct {
	ICoordinates
	location		string
}

func (c *coordinate) Equal(other ICoordinates) bool {
	return reflect.DeepEqual(c, other)
}

func (c *coordinate) Location() string {
	return c.location
}

func makeCoordinates(location string) ICoordinates {
	c := new(coordinate)
	c.location = location
	return c
}

type nodestub struct {
	INode
	ICommunicationHandler
	INodeFactory
	children		map[string]IChild
}

func (ns *nodestub) onMessageFromParent(msg IMessage) {
	response := MakeResponse(msg, nil)
	go ns.INode.send(ns.INode.Parent().ChildToParent(), response)
}

func (ns *nodestub) onMessageFromChild(msg IMessage) {

}

func (ns *nodestub) GetChild(coords ICoordinates) IChild {
	original, _ := coords.(ILocationCoordinates)
	child, ok := ns.children[original.Location()]
	if !ok {
		return nil
	}
	return child
}

func (ns *nodestub) makeChildNode(parentChannel IChild, childCoordinates ICoordinates) INode {
	child := makeNodeStub(parentChannel.Channel(), childCoordinates, ns.INode.Coordinates())
	loc, _ := childCoordinates.(ILocationCoordinates)
	ns.children[loc.Location()] = parentChannel
	return child
}

func makeNodeStub(parentChannel IChannel, coordinates, parentCoordinates ICoordinates) *nodestub {
	ns := new(nodestub)
	ns.INode = MakeNode(parentChannel, coordinates, parentCoordinates, ns, ns)
	ns.children = make(map[string]IChild)
	return ns
}

func TestCoordinatesEqual(t *testing.T) {
	c1 := makeCoordinates("test")
	c2 := makeCoordinates("test")
	if !c1.Equal(c2) {
		t.Error("Equality not evaluating correctly.")
	}

	c3 := makeCoordinates("test2")
	if c2.Equal(c3) {
		t.Error("Equality not evaluating correctly.")
	}
}

func TestMakeNewRootNode(t *testing.T) {
	coords := makeCoordinates("test")
	node := MakeNode(nil, coords, nil, nil, nil)
	if !node.Coordinates().Equal(coords) {
		t.Error("Node coordinates not set correctly.")
	}
}

func TestMakeNewNode(t *testing.T) {
	pcoords := makeCoordinates("parent")
	ccoords := makeCoordinates("child")
	ch := makeIChannel()
	MakeNode(ch, ccoords, pcoords, nil, nil)
	message := <- ch.ChildToParent()
	if message.Operation() != ChildInitialized {
		t.Error("Child not initialized properly")
	}
}

func TestRespondsToParent(t *testing.T) {
	pcoords := makeCoordinates("parent")
	ccoords := makeCoordinates("child")
	ch := makeIChannel()
	makeNodeStub(ch, ccoords, pcoords)
	<- ch.ChildToParent() // initialized
	cmd := MakeCommand("test", ccoords, pcoords, nil)
	ch.ParentToChild() <- cmd
	message := <- ch.ChildToParent()
	if message.MessageId() != cmd.MessageId() {
		t.Error("Child didn't echo parent.")
	}
}

func TestGetOrCreateChildChildDoesNotExist(t *testing.T) {
	pcoords := makeCoordinates("parent")
	ccoords := makeCoordinates("child")
	ch := makeIChannel()
	node := makeNodeStub(ch, ccoords, pcoords)
	<- ch.ChildToParent()
	coords := makeCoordinates("subChild")
	crCh := MakeMessageChannel()
	node.INode.GetOrCreateChild(crCh, coords)
	message := <- crCh
	if message.Operation() != ChildInitialized {
		t.Error("Child not created correctly.")
	}
}

func TestGetOrCreateChildExists(t *testing.T) {
	pcoords := makeCoordinates("parent")
	ccoords := makeCoordinates("child")
	ch := makeIChannel()
	node := makeNodeStub(ch, ccoords, pcoords)
	<- ch.ChildToParent()
	coords := makeCoordinates("subChild")
	crCh := MakeMessageChannel()
	node.INode.GetOrCreateChild(crCh, coords)
	<- crCh // child created
	node.INode.GetOrCreateChild(crCh, coords)
	<- crCh
	if len(node.children) != 1 {
		t.Error("Child not cached correctly.")
	}
}

func TestRouting(t *testing.T) {
	pcoords := makeCoordinates("parent")
	ccoords := makeCoordinates("child")
	ch := makeIChannel()
	node := makeNodeStub(ch, ccoords, pcoords)
	<- ch.ChildToParent() // initialized

	coords := makeCoordinates("subChild")
	crCh := MakeMessageChannel()
	node.INode.GetOrCreateChild(crCh, coords)
	<- crCh // child created

	ch.ParentToChild() <- MakeCommand("echo", coords, pcoords, nil)
	message := <- ch.ChildToParent()
	if message.Operation() != "echo" {
		t.Error("Message not echoed.")
	}
}

func TestRoutingBetweenChildren(t *testing.T) {
	pcoords := makeCoordinates("parent")
	ccoords := makeCoordinates("child")
	ch := makeIChannel()
	node := makeNodeStub(ch, ccoords, pcoords)
	<- ch.ChildToParent() // initialized
	node.children["test"] = makeIChild()
	node.children["test2"] = makeIChild()
	go node.INode.listenToChild(node.children["test"])
	node.children["test"].Channel().ChildToParent() <- MakeCommand("test", makeCoordinates("test2"), makeCoordinates("test"), nil)
	message := <- node.children["test2"].Channel().ParentToChild()
	if message.Operation() != "test" {
		t.Error("Message not routed properly between children.")
	}
}
