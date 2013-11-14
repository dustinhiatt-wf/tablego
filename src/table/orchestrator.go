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
}

func (o *orchestrator) onMessageFromParent(msg IMessage) {
	panic("An orchestrator should't get a message from a parent.")
}

func (o *orchestrator) onMessageFromChild(msg IMessage) {
	if msg.GetType() == node.Response {
		loc, _ := msg.SourceCoordinates().(ITableCoordinates)
		child, _ := o.children[loc.TableId()]
		go child.sendNotification(msg.MessageId(), msg)
	} else if msg.GetType() == node.Command {
		panic("Orchestrator got a command.")
	} else {
		//TODO: log error
	}
}

func (o *orchestrator) GetChild(coords ICoordinates) IChild {
	original, _ := coords.(ITableCoordinates)
	child, ok := o.children[original.TableId()]
	if !ok {
		return nil
	}
	return child
}

func (o *orchestrator) makeChildNode(parentChannel IChild, childCoordinates ICoordinates) INode {
	child := MakeTable(parentChannel, childCoordinates, o.INode.Coordinates())
	loc, _ := childCoordinates.(ITableCoordinates)
	o.children[loc.TableId()] = parentChannel
	return child
}

func MakeOrchestrator() *orchestrator {
	o := new(orchestrator)
	o.children = make(map[string]node.IChild)
	o.INode = node.MakeNode(nil, MakeCoordinates("", nil), nil, o, o)
	return o
}
