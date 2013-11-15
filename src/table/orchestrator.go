/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/5/13
 * Time: 3:36 PM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"node"
)

//var master = MakeOrchestrator()

type orchestrator struct {
	node.INode
	node.INodeFactory
	node.ICommunicationHandler
	children map[string]node.IChild
	collectionChannel						chan *addToChildMessage
}

func (o *orchestrator) OnMessageFromParent(msg node.IMessage) {
	panic("An orchestrator should't get a message from a parent.")
}

func (o *orchestrator) OnMessageFromChild(msg node.IMessage) {
	if msg.GetType() == node.Response {
		loc, _ := msg.SourceCoordinates().(ITableCoordinates)
		child, _ := o.children[loc.TableId()]
		go child.SendNotification(msg.MessageId(), msg)
	} else if msg.GetType() == node.Command {
		panic("Orchestrator got a command.")
	} else {
		//TODO: log error
	}
}

func (o *orchestrator) listenToCollection(ch chan *addToChildMessage) {
	ch <- nil
	for {
		select {
		case message := <- ch:
			if message.child == nil {
				child, ok := o.children[message.tableId]
				if !ok {
					message.returnChannel <- nil
					continue
				}
				message.returnChannel <- child
			} else {
				o.children[message.tableId] = message.child
				message.returnChannel <- message.child
			}
		}
	}
}

func (o *orchestrator) GetChild(coords node.ICoordinates) node.IChild {
	original, _ := coords.(ITableCoordinates)
	msg := makeAddToChildMessage(0, 0, original.TableId(), nil)
	o.collectionChannel <- msg
	rsp := <- msg.returnChannel
	return rsp
}

func (o *orchestrator) MakeChildNode(parentChannel node.IChild, childCoordinates node.ICoordinates) node.INode {
	child := MakeTable(parentChannel.Channel(), childCoordinates, o.INode.Coordinates())
	loc, _ := childCoordinates.(ITableCoordinates)
	msg := makeAddToChildMessage(0, 0, loc.TableId(), parentChannel)
	o.collectionChannel <- msg
	<- msg.returnChannel
	return child
}

func MakeOrchestrator() *orchestrator {
	o := new(orchestrator)
	o.children = make(map[string]node.IChild)
	ch := make(chan *addToChildMessage)
	o.collectionChannel = ch
	go o.listenToCollection(ch)
	<- ch //make sure the routine has started
	o.INode = node.MakeNode(nil, MakeCoordinates("", nil), nil, o, o)
	return o
}
