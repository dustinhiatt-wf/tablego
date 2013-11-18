package table

import (
	"node"
	"strconv"
)

type table struct {
	node.INode
	node.INodeFactory
	node.ICommunicationHandler
	children          map[int]map[int]node.IChild
	pendingRequests   map[string]chan node.IMessage
	collectionChannel chan *addToChildMessage
	requestChannel	  chan *addToPendingRequestsMessage
}

func (t *table) OnMessageFromParent(msg node.IMessage) {
	if msg.GetType() == node.Response {

	} else if msg.GetType() == node.Command {
		switch msg.Operation() {
		case GetValueRange:
			go t.getValueRangeByCellRange(t.INode.Parent().ChildToParent(), msg)
		case SubscribeToRange:
			go t.subscribeToRange(t.INode.Parent().ChildToParent(), msg)
		case UnsubscribeToRange:
			go t.unsubscribeToRange(t.INode.Parent().ChildToParent(), msg)
		default:
			resp := node.MakeResponse(msg, nil)
			go t.INode.Send(t.INode.Parent().ChildToParent(), resp)
		}
	} else {
		//TODO: log error
	}
}

func (t *table) DumpInformation(c node.ICoordinates) string {
	loc, _ := c.(ITableCoordinates)
	str := "Table ID: " + loc.TableId() + " ROW: " + strconv.Itoa(loc.CellLocation().Row()) + " COLUMN: " + strconv.Itoa(loc.CellLocation().Column())
	colok := false
	_, rowok := t.children[loc.CellLocation().Row()]
	if rowok {
		_, colok = t.children[loc.CellLocation().Row()][loc.CellLocation().Column()]
	}
	str = str + " VALUE AT THAT ROW: " + strconv.FormatBool(rowok) + " VALUE AT THAT COL: " + strconv.FormatBool(colok)
	return str
}

func (t *table) OnMessageFromChild(msg node.IMessage) {
	if msg.GetType() == node.Response {
		prMsg := makeAddToPendingRequestMessage(msg.MessageId(), nil)
		t.requestChannel <- prMsg
		ch := <- prMsg.returnChannel
		if ch != nil {
			t.INode.Send(ch, msg)
			//delete(t.pendingRequests, msg.MessageId())
			close(ch)
		}
	} else if msg.GetType() == node.Command {
		switch msg.Operation() {
		case SubscribeToRange:
			child := t.GetChild(msg.SourceCoordinates())
			go t.subscribeToRange(child.Channel().ParentToChild(), msg)
		case UnsubscribeToRange:
			child := t.GetChild(msg.SourceCoordinates())
			go t.unsubscribeToRange(child.Channel().ParentToChild(), msg)
		}
	} else {
		//TODO: log error
	}
}

func (t *table) GetChild(coords node.ICoordinates) node.IChild {
	original, _ := coords.(ITableCoordinates)
	msg := makeAddToChildMessage(original.CellLocation().Row(), original.CellLocation().Column(), "", nil)
	t.collectionChannel <- msg

	child := <-msg.returnChannel
	return child
}

func (t *table) listenToCollection(ch chan *addToChildMessage) {
	ch <- nil
	for {
		select {
		case message := <-ch:
			if message.child == nil {
				_, ok := t.children[message.row]
				if !ok {
					message.returnChannel <- nil
					continue
				}
				item, ok := t.children[message.row][message.column]
				if !ok {
					message.returnChannel <- nil
					continue
				}
				message.returnChannel <- item
			} else {
				_, ok := t.children[message.row]

				if !ok {
					t.children[message.row] = make(map[int]node.IChild)
				}
				t.children[message.row][message.column] = message.child
				message.returnChannel <- message.child
			}
		}
	}
}

func (t *table) listenToRequests(ch chan *addToPendingRequestsMessage) {
	ch <- nil
	for {
		select {
		case message := <-ch:
			if message.channel == nil {
				channel, ok := t.pendingRequests[message.id]
				if !ok {
					message.returnChannel <- nil
					continue
				}
				message.returnChannel <- channel
			} else {
				t.pendingRequests[message.id] = message.channel
				message.returnChannel <- message.channel
			}
		}
	}
}

func (t *table) MakeChildNode(parentChannel node.IChild, childCoordinates node.ICoordinates) node.INode {
	child := MakeCell(parentChannel.Channel(), childCoordinates, t.INode.Coordinates(), "")
	loc, _ := childCoordinates.(ITableCoordinates)

	msg := makeAddToChildMessage(loc.CellLocation().Row(), loc.CellLocation().Column(), "", parentChannel)
	t.collectionChannel <- msg
	<-msg.returnChannel
	close(msg.returnChannel)
	return child
}

func (t *table) getValueRangeByCellRange(ch chan node.IMessage, msg node.IMessage) {
	cr := MakeRangeFromBytes(msg.Payload())
	if cr == nil {
		go t.INode.Send(ch, node.MakeError(msg, nil))
	}
	go func() {
		loc, _ := t.INode.Coordinates().(ITableCoordinates)
		vr := new(valuerange)
		vr.Values = make(map[string]map[string]string)
		listeners := make([]chan node.IMessage, 0)
		for i := cr.StartRow; i < cr.StopRow; i++ {
			_, ok := t.children[i]
			if !ok {
				continue
			}
			vr.Values[strconv.Itoa(i)] = make(map[string]string)
			for j := cr.StartColumn; j < cr.StopColumn; j++ {
				child, ok := t.children[i][j]
				if ok {
					ch := node.MakeMessageChannel()
					cmd := node.MakeCommand(GetCellValue, MakeCoordinates(loc.TableId(), MakeCellLocation(i, j)), loc, nil)
					listeners = append(listeners, ch)
					t.pendingRequests[cmd.MessageId()] = ch
					vr.Values[strconv.Itoa(i)][strconv.Itoa(j)] = ""
					go t.INode.Send(child.Channel().ParentToChild(), cmd)
				}
			}
		}
		for _, ch := range listeners {
			message := <-ch
			cell := MakeCellFromBytes(message.Payload())
			loc, _ = message.SourceCoordinates().(ITableCoordinates)
			vr.Values[strconv.Itoa(loc.CellLocation().Row())][strconv.Itoa(loc.CellLocation().Column())] = cell.DisplayValue()
		}

		ch <- node.MakeResponse(msg, vr.ToBytes())
	}()
}

func (t *table) subscribeToRange(ch chan node.IMessage, msg node.IMessage) {
	cr := MakeRangeFromBytes(msg.Payload())
	if cr == nil {
		go t.INode.Send(ch, node.MakeError(msg, nil))
	}

	subscriberLoc := msg.SourceCoordinates().(ITableCoordinates)
	go func() {
		loc, _ := t.INode.Coordinates().(ITableCoordinates)
		vr := new(valuerange)
		vr.Values = make(map[string]map[string]string)
		valListeners := make([]chan node.IMessage, 0)
		for i := cr.StartRow; i < cr.StopRow; i++ {
			for j := cr.StartColumn; j < cr.StopColumn; j++ {
				if subscriberLoc.CellLocation() != nil {
					if subscriberLoc.CellLocation().Row() == i && subscriberLoc.CellLocation().Column() == j {
						continue
					}
				}

				ch := node.MakeMessageChannel()
				t.INode.GetOrCreateChild(ch, MakeCoordinates(loc.TableId(), MakeCellLocation(i, j)))
				<-ch //cell definitely exists now
				child := t.GetChild(MakeCoordinates(loc.TableId(), MakeCellLocation(i, j)))
				var sp *subscribePayload
				if subscriberLoc.CellLocation() != nil {
					sp = makeSubscribePayload(subscriberLoc.TableId(), subscriberLoc.CellLocation().Row(), subscriberLoc.CellLocation().Column(), true)
				} else {
					sp = makeSubscribePayload(subscriberLoc.TableId(), -1, -1, false)
				}
				cmd := node.MakeCommand(Subscribe, MakeCoordinates(loc.TableId(), MakeCellLocation(i, j)), t.INode.Coordinates(), sp.ToBytes())
				ch = node.MakeMessageChannel()
				req := makeAddToPendingRequestMessage(cmd.MessageId(), ch)
				t.requestChannel <-req
				<-req.returnChannel
				valListeners = append(valListeners, ch)
				go t.INode.Send(child.Channel().ParentToChild(), cmd)
			}
		}

		for _, ch := range valListeners {
			resp := <-ch
			cellLoc, _ := resp.SourceCoordinates().(ITableCoordinates)
			_, ok := vr.Values[strconv.Itoa(cellLoc.CellLocation().Row())]
			if !ok {
				vr.Values[strconv.Itoa(cellLoc.CellLocation().Row())] = make(map[string]string)
			}
			cell := MakeCellFromBytes(resp.Payload())
			vr.Values[strconv.Itoa(cellLoc.CellLocation().Row())][strconv.Itoa(cellLoc.CellLocation().Column())] = cell.DisplayValue()
		}
		ch <- node.MakeResponse(msg, vr.ToBytes())
	}()
}

func (t *table) unsubscribeToRange(ch chan node.IMessage, msg node.IMessage) {
	cr := MakeRangeFromBytes(msg.Payload())
	if cr == nil {
		go t.INode.Send(ch, node.MakeError(msg, nil))
		return
	}

	subscriberLoc := msg.SourceCoordinates().(ITableCoordinates)

	go func() {
		loc, _ := t.INode.Coordinates().(ITableCoordinates)
		listeners := make([]chan node.IMessage, 0)
		for i := cr.StartRow; i < cr.StopRow; i++ {
			for j := cr.StartColumn; j < cr.StopColumn; j++ {
				msg := makeAddToChildMessage(i, j, "", nil)
				t.collectionChannel <- msg
				child := <- msg.returnChannel
				if child != nil {
					ch := node.MakeMessageChannel()
					var sp *subscribePayload
					if subscriberLoc.CellLocation() != nil {
						sp = makeSubscribePayload(subscriberLoc.TableId(), subscriberLoc.CellLocation().Row(), subscriberLoc.CellLocation().Column(), true)
					} else {
						sp = makeSubscribePayload(subscriberLoc.TableId(), -1, -1, false)
					}
					cmd := node.MakeCommand(Unsubscribe, MakeCoordinates(loc.TableId(), MakeCellLocation(i, j)), t.INode.Coordinates(), sp.ToBytes())
					ch = node.MakeMessageChannel()
					req := makeAddToPendingRequestMessage(cmd.MessageId(), ch)
					t.requestChannel <-req
					<-req.returnChannel
					listeners = append(listeners, ch)
					go t.INode.Send(child.Channel().ParentToChild(), cmd)
				}
			}
		}

		for _, ch := range listeners {
			<- ch
		}

		ch <- node.MakeResponse(msg, nil)
	}()
}

func MakeTable(parentChannel node.IChannel, coordinates, parentCoordinates node.ICoordinates) *table {
	t := new(table)
	t.children = make(map[int]map[int]node.IChild)
	t.pendingRequests = make(map[string]chan node.IMessage)
	ch := make(chan *addToChildMessage)
	t.collectionChannel = ch
	go t.listenToCollection(t.collectionChannel)
	<-t.collectionChannel
	reqCh := make(chan *addToPendingRequestsMessage)
	t.requestChannel = reqCh
	go t.listenToRequests(t.requestChannel)
	<-t.requestChannel
	// this is where we need to load and parse
	t.INode = node.MakeNode(parentChannel, coordinates, parentCoordinates, t, t)
	t.INode.Initialize()
	return t
}
