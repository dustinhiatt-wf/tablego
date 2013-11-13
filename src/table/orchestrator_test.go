/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/5/13
 * Time: 3:54 PM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"testing"
)

func TestOrchestratorCachesTable(t *testing.T) {
	o := MakeOrchestrator()
	ch := MakeMessageChannel()
	o.GetTableById("test", ch)
	if len(o.tables) != 1 {
		t.Error("Table not cached correctly.")
	}
	o.GetTableById("test", ch)
	if len(o.tables) != 1 {
		t.Error("Table cached twice.")
	}
}

func TestOrchestratorCanRemoveTable(t *testing.T) {
	o := MakeOrchestrator()
	ch := MakeMessageChannel()
	o.GetTableById("test", ch)
	removeTable(o, "test")
	_, ok := o.tables["test"]
	if ok == true {
		t.Error("Table not removed correctly.")
	}
}

func TestCheckTableLoaded(t *testing.T) {
	o := MakeOrchestrator()
	ch := MakeMessageChannel()
	if o.IsTableLoaded("test") {
		t.Error("Table not correctly identified as not being loaded.")
	}
	o.GetTableById("test", ch)
	if !o.IsTableLoaded("test") {
		t.Error("Table not correctly identified as being loaded.")
	}
}

func TestOrchestratorNotifiesOfTableOpen(t *testing.T) {
	o := MakeOrchestrator()
	ch := MakeMessageChannel()
	o.GetTableById("test", ch)
	msg := <- ch
	if msg.Operation() != TableOpened {
		t.Error("Table not opened correctly.")
	} else if msg.SourceTable() != "test" {
		t.Error("Table opened source not opened correctly.")
	}
}

func TestOrchestratorNotifiesMultipleUsersOfTableOpen(t *testing.T) {
	o := MakeOrchestrator()
	chOne := MakeMessageChannel()
	chTwo := MakeMessageChannel()
	o.GetTableById("test", chOne)
	o.GetTableById("test", chTwo)
	if len(o.tables["test"].pendingRequests["tableOpened"].subscribers) != 2 {
		t.Error("Simultaneous gets not handled correctly.")
	}
	messageOne := <- chOne
	messageTwo := <- chTwo

	if messageOne.SourceTable() != "test" || messageTwo.SourceTable() != "test" {
		t.Error("Table entity not received correctly.")
	}
}

func TestOrchestratorNotifiesUsersWhoGetTableAfterLoad(t *testing.T) {
	o := MakeOrchestrator()
	chOne := MakeMessageChannel()
	o.GetTableById("test", chOne)
	<- chOne
	ch := MakeMessageChannel()
	o.GetTableById("test", ch)
	if len(o.tables["test"].pendingRequests[TableOpened].subscribers) != 0 {
		t.Error("Subscribed to an already loaded table.")
	}
	message := <- ch
	if message.SourceTable() != "test" {
		t.Error("Table not returned correctly.")
	}
}

func TestOrchestratorLoadsTableWhenCommandReceivedForTableNotLoaded(t *testing.T) {
	o := MakeOrchestrator()
	ch := MakeMessageChannel()
	cmd := MakeCommand("test", "test", "", nil, nil, nil)
	o.sendCommand(cmd, ch)
	message := <- ch
	if message.MessageId() != cmd.MessageId() {
		t.Error("Response not formulated correctly.")
	}
}

func TestOrchestratorRespondsToMultipleTables(t *testing.T) {
	o := MakeOrchestrator()
	chOne := MakeMessageChannel()
	chTwo := MakeMessageChannel()
	cmdOne := MakeCommand("test", "test", "", nil, nil, nil)
	cmdTwo := MakeCommand("test", "test2", "", nil, nil, nil)
	o.sendCommand(cmdOne, chOne)
	o.sendCommand(cmdTwo, chTwo)
	messageOne := <- chOne
	messageTwo := <- chTwo

	if messageOne.MessageId() != cmdOne.MessageId() {
		t.Error("Response one not created correctly.")
	}
	if messageTwo.MessageId() != cmdTwo.MessageId() {
		t.Error("Response two not created correctly.")
	}
}

func TestOrchestratorRespondsToMultipleRequestsSameTable(t *testing.T) {
	o := MakeOrchestrator()
	chOne := MakeMessageChannel()
	chTwo := MakeMessageChannel()
	cmdOne := MakeCommand("test", "test", "", nil, nil, nil)
	cmdTwo := MakeCommand("test", "test", "", nil, nil, nil)
	o.sendCommand(cmdOne, chOne)
	o.sendCommand(cmdTwo, chTwo)
	messageOne := <- chOne
	messageTwo := <- chTwo

	if messageOne.MessageId() != cmdOne.MessageId() {
		t.Error("Response one not created correctly.")
	}
	if messageTwo.MessageId() != cmdTwo.MessageId() {
		t.Error("Response two not created correctly.")
	}

	if len(o.tables) != 1 {
		t.Error("Tables not cached correctly.")
	}
}

/*
This will deadlock if not forwarded
 */
func TestOrchestratorForwarding(t *testing.T) {
	o := MakeOrchestrator()
	ch := MakeMessageChannel()
	o.GetTableById("test", ch)
	<- ch
	o.GetTableById("test2", ch)
	<- ch
	o.tables["test2"].channel = MakeTableOrchestratorChannel()
	o.tables["test"].channel.tableToOrchestrator <- MakeCommand("test", "test2", "test", nil, nil, nil)
	<- o.tables["test2"].channel.orchestratorToTable
}

func TestSubscribeToDistantTableIntegration(t *testing.T) {
	o := MakeOrchestrator()
	ch := MakeMessageChannel()
	orCh := MakeTableChannel()
	tbl := MakeTable("test", orCh.channel)
	<- orCh.channel.tableToOrchestrator
	o.tables["test"] = orCh
	orCh.tableInitialized = true
	go o.listenToTable(orCh)

	cc := MakeCellChannel()
	MakeCell(1, 1, "", MakeCellChannel())
	tbl.cells[1] = make(map[int]*cellChannel)
	tbl.cells[1][1] = cc
	cc.cellInitialized = true
	go tbl.listenToCell(cc)
	o.GetTableById("test2", ch)
	<- ch // table created

	o.sendCommand(MakeCommand(EditCellValue, "test2", "", MakeCellLocation(1, 1), nil, MakeTableCommand("test").ToBytes()), ch)
	message := <-ch
	cc.channel.cellToTable <- MakeCommand(Subscribe, "test2", "test", MakeCellLocation(1, 1), MakeCellLocation(1, 1), nil)
	message = <- cc.channel.tableToCell // we are subscribed
	c := MakeCellFromBytes(message.Payload())
	if c.CellDisplayValue != "test" {
		t.Error("Subscribe payload not returned correctly.")
	}
	o.sendCommand(MakeCommand(EditCellValue, "test2", "", MakeCellLocation(1, 1), nil, MakeTableCommand("test2").ToBytes()), ch)
	<- ch

	message = <- cc.channel.tableToCell
	c = MakeCellFromBytes(message.Payload())
	if c.DisplayValue() != "test2" {
		t.Error("Subscribe payload after update not returned correctly.")
	}
}
