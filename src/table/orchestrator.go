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
	children          map[string]node.IChild
	collectionChannel chan *addToChildMessage
}

func (o *orchestrator) OnMessageFromParent(msg node.IMessage) {
	panic("An orchestrator should't get a message from a parent.")
}

func (o *orchestrator) OnMessageFromChild(msg node.IMessage) {
	if msg.GetType() == node.Response {
		loc, _ := msg.SourceCoordinates().(ITableCoordinates)
		child, _ := o.children[loc.TableId()]
		go child.SendNotification(msg.MessageId(), msg, true)
	} else if msg.GetType() == node.Command {
		if msg.Operation() == CellUpdated {
			loc := msg.SourceCoordinates().(ITableCoordinates)
			child := o.GetChild(MakeCoordinates(loc.TableId(), nil))
			child.SendNotification(CellUpdated, msg, false)
		}
	} else {
		//TODO: log error
	}
}

func (o *orchestrator) sendCommand(msg node.IMessage, observer chan node.IMessage) {
	go func () {
		ch := node.MakeMessageChannel()
		o.INode.GetOrCreateChild(ch, msg.TargetCoordinates())
		<- ch
		child := o.GetChild(msg.TargetCoordinates())
		child.Subscribe(msg.MessageId(), observer)
		go o.INode.Send(child.Channel().ParentToChild(), msg)
	}()
}

func (o *orchestrator) subscribeToTableRange(cellRange *cellrange, values, observer chan node.IMessage) {
	if cellRange.TableId == "" {
		go func () {
			observer <- nil
			close(observer)
		}()
	}
	go func () {
		ch := node.MakeMessageChannel()
		o.INode.GetOrCreateChild(ch, MakeCoordinates(cellRange.TableId, nil))
		<-ch // table exists now
		child := o.GetChild(MakeCoordinates(cellRange.TableId, nil))
		cmd := node.MakeCommand(SubscribeToRange, MakeCoordinates(cellRange.TableId, nil), MakeCoordinates("", nil), cellRange.ToBytes())
		child.Subscribe(CellUpdated, observer)
		child.Subscribe(cmd.MessageId(), ch)
		go o.INode.Send(child.Channel().ParentToChild(), cmd)
		msg := <- ch
		values <- msg
	}()
}

func (o *orchestrator) listenToCollection(ch chan *addToChildMessage) {
	ch <- nil
	for {
		select {
		case message := <-ch:
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
	rsp := <-msg.returnChannel
	return rsp
}

func (o *orchestrator) MakeChildNode(parentChannel node.IChild, childCoordinates node.ICoordinates) node.INode {
	loc, _ := childCoordinates.(ITableCoordinates)
	child := MakeTable(parentChannel.Channel(), MakeCoordinates(loc.TableId(), nil), o.INode.Coordinates())

	msg := makeAddToChildMessage(0, 0, loc.TableId(), parentChannel)
	o.collectionChannel <- msg
	<-msg.returnChannel
	return child
}

func MakeOrchestrator() *orchestrator {
	o := new(orchestrator)
	o.children = make(map[string]node.IChild)
	ch := make(chan *addToChildMessage)
	o.collectionChannel = ch
	go o.listenToCollection(ch)
	<-ch //make sure the routine has started
	o.INode = node.MakeNode(nil, MakeCoordinates("", nil), nil, o, o)
	o.INode.Initialize()
	return o
}
