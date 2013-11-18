package table

import (
	"node"
	"sync"
)

var master = MakeOrchestrator()

/*
For unit testing
 */

func resetMaster() {
	master = MakeOrchestrator()
}

type orchestrator struct {
	node.INode
	node.INodeFactory
	node.ICommunicationHandler
	children          	map[string]node.IChild
	collectionChannel 	chan *addToChildMessage
	childMutex 			sync.Mutex
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

func (o *orchestrator) GetChild(coords node.ICoordinates) node.IChild {
	original, _ := coords.(ITableCoordinates)
	o.childMutex.Lock()
	child, ok := o.children[original.TableId()]
	o.childMutex.Unlock()
	if ok {
		return child
	}
	return nil
}

func (o *orchestrator) MakeChildNode(parentChannel node.IChild, childCoordinates node.ICoordinates) node.INode {
	loc, _ := childCoordinates.(ITableCoordinates)
	child := MakeTable(parentChannel.Channel(), MakeCoordinates(loc.TableId(), nil), o.INode.Coordinates())

	o.childMutex.Lock()
	o.children[loc.TableId()] = parentChannel
	o.childMutex.Unlock()
	return child
}

func MakeOrchestrator() *orchestrator {
	o := new(orchestrator)
	o.children = make(map[string]node.IChild)
	o.INode = node.MakeNode(nil, MakeCoordinates("", nil), nil, o, o)
	o.INode.Initialize()
	return o
}

func SubscribeToTableRange(cellRange *cellrange, values, observer chan node.IMessage) {
	master.subscribeToTableRange(cellRange, values, observer)
}

func UpdateCellAtLocation(tableId string, row, column int, value string) {
	cmd := node.MakeCommand(EditCellValue, MakeCoordinates(tableId, MakeCellLocation(row, column)), MakeCoordinates("", nil), MakeTableCommand(value).ToBytes())
	ch := node.MakeMessageChannel()
	master.sendCommand(cmd, ch)
	<- ch
	close(ch)
}
