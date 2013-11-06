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
	ch := MakeValueChannel()
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
	ch := MakeValueChannel()
	o.GetTableById("test", ch)
	removeTable(o, "test")
	_, ok := o.tables["test"]
	if ok == true {
		t.Error("Table not removed correctly.")
	}
}

func TestCheckTableLoaded(t *testing.T) {
	o := MakeOrchestrator()
	ch := MakeValueChannel()
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
	ch := MakeValueChannel()
	o.GetTableById("test", ch)
	msg := <- ch
	if msg.table.id != "test" {
		t.Error("Table not created correctly.")
	}
}

func TestOrchestratorNotifiesMultipleUsersOfTableOpen(t *testing.T) {
	o := MakeOrchestrator()
	chOne := MakeValueChannel()
	chTwo := MakeValueChannel()
	o.GetTableById("test", chOne)
	o.GetTableById("test", chTwo)
	if len(o.tables["test"].pendingRequests["tableOpened"].subscribers) != 2 {
		t.Error("Simultaneous gets not handled correctly.")
	}
	messageOne := <- chOne
	messageTwo := <- chTwo

	if messageOne.table.id != "test" || messageTwo.table.id != "test" {
		t.Error("Table entity not received correctly.")
	}
}

func TestOrchestratorNotifiesUsersWhoGetTableAfterLoad(t *testing.T) {
	o := MakeOrchestrator()
	chOne := MakeValueChannel()
	o.GetTableById("test", chOne)
	<- chOne
	ch := MakeValueChannel()
	o.GetTableById("test", ch)
	if len(o.tables["test"].pendingRequests[TableOpened].subscribers) != 0 {
		t.Error("Subscribed to an already loaded table.")
	}
	message := <- ch
	if message.table.id != "test" {
		t.Error("Table not returned correctly.")
	}
}
