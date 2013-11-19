package table

import (
	"node"
	"testing"
)

func TestCreateOrchestrator(t *testing.T) {
	o := MakeOrchestrator()
	if o.INode.Parent() != nil {
		t.Error("Orchestrator created with parent.")
	}
	coords := MakeCoordinates("", nil)

	if !o.INode.Coordinates().Equal(coords) {
		t.Error("Orchestrator created with wrong coordinates.")
	}
	if o.INode.ParentCoordinates() != nil {
		t.Error("Orchestrator created with parent coordinates.")
	}
}

func TestCreateChildTable(t *testing.T) {
	o := MakeOrchestrator()
	ch := node.MakeMessageChannel()
	o.INode.GetOrCreateChild(ch, MakeCoordinates("test", nil))
	<-ch //table initialized
	if len(o.children) != 1 {
		t.Error("Child table not created.")
	}
}

func TestOrchestratorReturnsCachedTable(t *testing.T) {
	o := MakeOrchestrator()
	ch := node.MakeMessageChannel()
	o.INode.GetOrCreateChild(ch, MakeCoordinates("test", nil))
	<-ch
	o.INode.GetOrCreateChild(ch, MakeCoordinates("test", nil))
	<-ch
	if len(o.children) != 1 {
		t.Error("Child cached more than once")
	}
}

func TestOrchestratorNotifiesOnTableCreation(t *testing.T) {
	o := MakeOrchestrator()
	ch := node.MakeMessageChannel()
	chTwo := node.MakeMessageChannel()
	o.INode.GetOrCreateChild(ch, MakeCoordinates("test", nil))
	o.INode.GetOrCreateChild(chTwo, MakeCoordinates("test", nil))
	<-ch
	<-chTwo
	if len(o.children) != 1 {
		t.Error("Child tables not created correctly.")
	}
}

func TestOrchestratorRoutesToChild(t *testing.T) {
	o := MakeOrchestrator()
	ch := node.MakeMessageChannel()
	o.INode.GetOrCreateChild(ch, MakeCoordinates("test", nil))
	<-ch
	tblCh := node.MakeIChild()
	o.children["test2"] = tblCh
	o.children["test"].Channel().ChildToParent() <- node.MakeCommand("test", MakeCoordinates("test2", nil), MakeCoordinates("test", nil), nil)
	msg := <-tblCh.Channel().ParentToChild()
	if msg.Operation() != "test" {
		t.Error("Orchestrator did not forward message correctly.")
	}
}

func TestOrchestratorSendCommand(t *testing.T) {
	o := MakeOrchestrator()
	cmd := node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes())
	ch := node.MakeMessageChannel()
	o.sendCommand(cmd, ch)
	msg := <- ch
	if msg.MessageId() != cmd.MessageId() {
		t.Error("Cell not created correctly.")
	}
	cmd = node.MakeCommand(GetCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), nil)
	o.sendCommand(cmd, ch)
	msg = <- ch
	c := MakeCellFromBytes(msg.Payload())
	if c.DisplayValue() != "test" {
		t.Error("Cell value not retrieved correctly.")
	}
}

func TestSubscribeToOrchestrator(t *testing.T) {
	o := MakeOrchestrator()
	cr := MakeRange("A1:C3")
	cr.TableId = "test"
	values := node.MakeMessageChannel()
	notifications := node.MakeMessageChannel()

	o.subscribeToTableRange(cr, values, notifications)
	msg := <- values

	vr := MakeValueRangeFromBytes(msg.Payload())
	if len(vr.Values) != 3 {
		t.Error("Value range not returned correctly.")
	}

	ch := node.MakeMessageChannel()
	o.sendCommand(node.MakeCommand(EditCellValue, MakeCoordinates("test", MakeCellLocation(1, 1)), MakeCoordinates("", nil), MakeTableCommand("test").ToBytes()), ch)
	<-ch // cell updated

	msg = <- notifications

	c := MakeCellFromBytes(msg.Payload())
	if c.DisplayValue() != "test" {
		t.Error("Orchestrator not notified correctly.")
	}
}

func TestSubscriberWhenCellGetsUpdated(t *testing.T) {
	resetMaster()
	cr := MakeRange("A1:B3")
	cr.TableId = "test"
	values, observer := node.MakeMessageChannel(), node.MakeMessageChannel()
	SubscribeToTableRange(cr, values, observer)
	<- values // we got our range and subscribed
	UpdateCellAtLocation("test", 1, 1, "1")
	msg := <- observer
	cell := MakeCellFromBytes(msg.Payload())
	if cell.DisplayValue() != "1" {
		t.Error("Sum not calculated correctly.")
	}
}

func TestCreateFormulaOverRange(t *testing.T) {
	resetMaster()
	cr := MakeRange("A1:C5")
	cr.TableId = "test"
	values, observer := node.MakeMessageChannel(), node.MakeMessageChannel()
	SubscribeToTableRange(cr, values, observer)
	<- values // we got our range and subscribed
	UpdateCellAtLocation("test", 1, 1, "1")
	<- observer // we are updated with 1 value

	UpdateCellAtLocation("test", 3, 0, "=sum(A1:B2)")

	msg := <- observer
	c := MakeCellFromBytes(msg.Payload())
	if c.DisplayValue() != "1" {
		t.Error("Formula not calculated correctly.")
	}
}

func TestFormulaValueGetsUpdatedWhenRangeUpdated(t *testing.T) {
	resetMaster()
	cr := MakeRange("A1:B4")
	cr.TableId = "test"
	values, observer := node.MakeMessageChannel(), node.MakeMessageChannel()
	SubscribeToTableRange(cr, values, observer)
	<- values // we got our range and subscribed
	UpdateCellAtLocation("test", 1, 1, "1")
	<- observer // we are updated with 1 value

	UpdateCellAtLocation("test", 3, 0, "=sum(A1:B2)")
	<- observer //formula now correct

	UpdateCellAtLocation("test", 1, 1, "30")
	var message node.IMessage
	msg := <- observer
	if msg.SourceCoordinates().(ITableCoordinates).CellLocation().Row() == 3 {
		message = msg
	}
	msg = <- observer
	if msg.SourceCoordinates().(ITableCoordinates).CellLocation().Row() == 3 {
		message = msg
	}

	c := MakeCellFromBytes(message.Payload())
	if c.DisplayValue() != "30" {
		t.Error("Formula value did not update correctly.")
	}
}

func TestUpdateCellValue(t *testing.T) {
	resetMaster()
	cr := MakeRange("A1:B4")
	cr.TableId = "test"
	values, observer := node.MakeMessageChannel(), node.MakeMessageChannel()
	SubscribeToTableRange(cr, values, observer)
	<- values // we got our range and subscribed

	UpdateCellAtLocation("test", 1, 1, "`")
	<- observer // we are updated with 1 value

	UpdateCellAtLocation("test", 1, 1, "1")
	msg := <- observer

	cell := MakeCellFromBytes(msg.Payload())
	if cell.DisplayValue() != "1" {
		t.Error("Cell not updated correctly.")
	}
}

func TestTextGetsUpdatedToNumberInFormula(t *testing.T) {
	resetMaster()
	cr := MakeRange("A1:B4")
	cr.TableId = "test"
	values, observer := node.MakeMessageChannel(), node.MakeMessageChannel()
	SubscribeToTableRange(cr, values, observer)
	<- values // we got our range and subscribed
	UpdateCellAtLocation("test", 1, 1, "test")
	<- observer // we are updated with 1 value

	UpdateCellAtLocation("test", 3, 0, "=sum(A1:B2)")
	msg := <- observer //formula now correct
	c := MakeCellFromBytes(msg.Payload())
	if c.DisplayValue() != "0" {
		t.Error("Sum not calculated correctly.")
	}

	UpdateCellAtLocation("test", 1, 1, "1")
	var message node.IMessage
	msg = <- observer
	if msg.SourceCoordinates().(ITableCoordinates).CellLocation().Row() == 3 {
		message = msg
	}
	msg = <- observer
	if msg.SourceCoordinates().(ITableCoordinates).CellLocation().Row() == 3 {
		message = msg
	}

	c = MakeCellFromBytes(message.Payload())
	if c.DisplayValue() != "1" {
		t.Error("Formula not updated correctly.")
	}
}

func TestCrossTableFormula(t *testing.T) {
	resetMaster()
	cr := MakeRange("A1:B4")
	cr.TableId = "test1"
	values, observer := node.MakeMessageChannel(), node.MakeMessageChannel()
	SubscribeToTableRange(cr, values, observer)
	<- values // we got our range and subscribed

	observerTwo := node.MakeMessageChannel()
	cr = MakeRange("test2!A1:B4")
	SubscribeToTableRange(cr, values, observerTwo)
	<- values //have range on second table

	UpdateCellAtLocation("test1", 1, 1, "1")
	<- observer // we are updated with 1 value

	UpdateCellAtLocation("test2", 3, 0, "=sum(test1!A1:B2)")
	msg := <- observerTwo //formula now correct

	c := MakeCellFromBytes(msg.Payload())
	if c.DisplayValue() != "1" {
		t.Error("Sum not calculated correctly.")
	}
}

func TestCrossTableFormulaWithUpdate(t *testing.T) {
	resetMaster()
	cr := MakeRange("A1:B4")
	cr.TableId = "test1"
	values, observer := node.MakeMessageChannel(), node.MakeMessageChannel()
	SubscribeToTableRange(cr, values, observer)
	<- values // we got our range and subscribed

	observerTwo := node.MakeMessageChannel()
	cr = MakeRange("test2!A1:B4")
	SubscribeToTableRange(cr, values, observerTwo)
	<- values //have range on second table

	UpdateCellAtLocation("test1", 1, 1, "1")
	<- observer // we are updated with 1 value

	UpdateCellAtLocation("test2", 3, 0, "=sum(test1!A1:B2)")
	<- observerTwo //formula now correct

	UpdateCellAtLocation("test1", 1, 1, "2")
	<- observer // target cell updated

	msg := <- observerTwo
	c := MakeCellFromBytes(msg.Payload())
	if c.DisplayValue() != "2" {
		t.Error("Formula cell didn't update after target range update.")
	}
}
