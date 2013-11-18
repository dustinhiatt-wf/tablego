package table

import (
	"encoding/json"
	"node"
	"strings"
	"reflect"
)

type ICell interface {
	DisplayValue() string
	SetValue(value string, timestamp int)
	GetCellValue() *cellValue
}

type cell struct {
	ISerializable
	ICell
	node.INode
	node.ICommunicationHandler
	node.INodeFactory
	CellDisplayValue 					string
	Value            					string
	LastUpdated      					int
	observers        					*observers
	pendingRequests  					map[string]chan node.IMessage
	isFormula		 					bool
	formulaValueRange					*valuerange
	formulaRange						*cellrange
	requestChannel	  					chan *addToPendingRequestsMessage
}

type cellValue struct {
	CellDisplayValue string
	Value            string
	LastUpdated      int
}

func makeCellValue(cellDisplayValue, value string, lastUpdated int) *cellValue {
	cv := new(cellValue)
	cv.CellDisplayValue = cellDisplayValue
	cv.Value = value
	cv.LastUpdated = lastUpdated
	return cv
}

func (c *cell) GetCellValue() *cellValue {
	return makeCellValue(c.DisplayValue(), c.Value, c.LastUpdated)
}

func (c *cell) ToBytes() []byte {
	cv := c.GetCellValue()
	res, err := json.Marshal(cv)
	if err != nil {
		return nil
	}

	return res
}

func MakeCellFromBytes(bytes []byte) *cell {
	var m cellValue
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return nil
	}
	cv := &m
	cell := new(cell)
	cell.CellDisplayValue = cv.CellDisplayValue
	cell.Value = cv.Value
	cell.LastUpdated = cv.LastUpdated
	return cell
}

func (c *cell) DisplayValue() string {
	return c.CellDisplayValue
}

func (c *cell) SetValue(value string, timestamp int) {
	if value == c.Value {
		return
	}
	c.Value = value
	c.LastUpdated = timestamp
	c.CellDisplayValue = c.parseValue(value)
	c.observers.notifyObservers(CellUpdated, c.INode.Parent().ChildToParent(), c.ToBytes(), c.INode.Coordinates())
}

func (c *cell) OnMessageFromParent(msg node.IMessage) {
	if msg.GetType() == node.Response {
		ch, ok := c.pendingRequests[msg.MessageId()]
		if ok {
			c.INode.Send(ch, msg)
			delete(c.pendingRequests, msg.MessageId())
			close(ch)
		}
	} else if msg.GetType() == node.Command {
		switch msg.Operation() {
		case GetCellValue:
			resp := node.MakeResponse(msg, c.ToBytes())
			go c.INode.Send(c.INode.Parent().ChildToParent(), resp)
		case EditCellValue:
			if msg.Timestamp() < c.LastUpdated {
				err := node.MakeError(msg, nil)
				go c.Send(c.INode.Parent().ChildToParent(), err)
				return
			}

			tblCmd := MakeTableCommandFromJson(msg.Payload())
			c.SetValue(tblCmd.Value, msg.Timestamp())
			resp := node.MakeResponse(msg, c.ToBytes())
			go c.Send(c.INode.Parent().ChildToParent(), resp)
		case Subscribe:
			sp := makeSubscribePayloadFromBytes(msg.Payload())
			c.Subscribe(sp)
			resp := node.MakeResponse(msg, c.ToBytes())
			go c.Send(c.INode.Parent().ChildToParent(), resp)
		case Unsubscribe:
			sp := makeSubscribePayloadFromBytes(msg.Payload())
			c.Unsubscribe(sp)
			resp := node.MakeResponse(msg, c.ToBytes())
			go c.Send(c.INode.Parent().ChildToParent(), resp)
		case CellUpdated:
			loc, _ := msg.SourceCoordinates().(ITableCoordinates)
			cell := MakeCellFromBytes(msg.Payload())
			c.formulaValueRange.update(loc.CellLocation().Row(), loc.CellLocation().Column(), cell.DisplayValue())
			result := c.executeFormula()
			c.LastUpdated = msg.Timestamp()
			c.CellDisplayValue = result
			c.observers.notifyObservers(CellUpdated, c.INode.Parent().ChildToParent(), c.ToBytes(), c.INode.Coordinates())
		default:
			resp := node.MakeResponse(msg, nil)
			c.INode.Send(c.INode.Parent().ChildToParent(), resp)
		}

	} else {
		//TODO: log error
	}
}

func (c *cell) listenToRequests(ch chan *addToPendingRequestsMessage) {
	ch <- nil
	for {
		select {
		case message := <-ch:
			if message.channel == nil {
				channel, ok := c.pendingRequests[message.id]
				if !ok {
					message.returnChannel <- nil
					continue
				}
				message.returnChannel <- channel
			} else {
				c.pendingRequests[message.id] = message.channel
				message.returnChannel <- message.channel
			}
		}
	}
}

func (c *cell) OnMessageFromChild(msg node.IMessage) {
	panic("Cells should not receive messages from children")
}

func (c *cell) GetChild(coords node.ICoordinates) node.IChild {
	//TODO: so we can handle embedded formulas, they should be children
	panic("Cells have no children currently.")
}

func (c *cell) makeChildNode(parentChannel node.IChild, childCoordinates node.ICoordinates) node.INode {
	panic("Cells can't create children.")
}

func (c *cell) Subscribe(sp *subscribePayload) {
	c.observers.addObserver(sp)
}

func (c *cell) Unsubscribe(sp *subscribePayload) {
	c.observers.removeObserver(sp)
}

func (c *cell) unsubscribeCellRange() {
	if c.formulaRange == nil {
		return
	}

	cmd := node.MakeCommand(UnsubscribeToRange, MakeCoordinates(c.formulaRange.TableId, nil), c.INode.Coordinates(), c.formulaRange.ToBytes())
	go c.INode.Send(c.INode.Parent().ChildToParent(), cmd)
}

func (c *cell) sum(parts []string) string {
	cr := MakeRange(parts[0])
	loc, _ := c.INode.Coordinates().(ITableCoordinates)
	if cr.TableId == "" {
		cr.TableId = loc.TableId()
	}
	if !reflect.DeepEqual(cr, c.formulaRange) {
		c.unsubscribeCellRange()
		c.formulaRange = cr
		cmd := node.MakeCommand(SubscribeToRange, MakeCoordinates(cr.TableId, nil), c.INode.Coordinates(), cr.ToBytes())
		ch := node.MakeMessageChannel()
		msg := makeAddToPendingRequestMessage(cmd.MessageId(), ch)
		c.requestChannel <- msg
		<- msg.returnChannel // waiting channel added
		go c.INode.Send(c.INode.Parent().ChildToParent(), cmd)
		resp := <- ch // we got our value range

		vr := MakeValueRangeFromBytes(resp.Payload())
		c.formulaValueRange = vr
	}

	return c.formulaValueRange.Sum()
}

func (c *cell) executeFormula() string {
	if !strings.HasPrefix(c.Value, "=") {
		return ""
	}

	parts := parseFormula(c.Value)
	switch strings.ToLower(parts[0]) {
	case "sum":
		return c.formulaValueRange.Sum()
	default:
		return ""
	}
}

func (c *cell) parseValue(value string) string {
	if !strings.HasPrefix(value, "=") {
		c.isFormula = false
		return value
	}
	parts := parseFormula(value)
	c.isFormula = true
	result := value
	switch strings.ToLower(parts[0]) {
	case "sum":
		result = c.sum(parts[1:])
	}

	return result
}

func MakeCell(parentChannel node.IChannel, coordinates, parentCoordinates node.ICoordinates, value string) *cell {
	cell := new(cell)
	cell.Value = value
	cell.observers = MakeObservers()
	cell.pendingRequests = make(map[string]chan node.IMessage)
	cell.INode = node.MakeNode(parentChannel, coordinates, parentCoordinates, cell, cell)

	reqCh := make(chan *addToPendingRequestsMessage)
	cell.requestChannel = reqCh
	go cell.listenToRequests(cell.requestChannel)
	<-cell.requestChannel
	cell.CellDisplayValue = cell.parseValue(value)
	cell.INode.Initialize()
	return cell
}
