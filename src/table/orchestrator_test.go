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
	o.GetTableById("test")
	if len(o.tables) != 1 {
		t.Error("Table not cached correctly.")
	}
	o.GetTableById("test")
	if len(o.tables) != 1 {
		t.Error("Table cached twice.")
	}
}

func TestOrchestratorCanRemoveTable(t *testing.T) {
	o := MakeOrchestrator()
	table := o.GetTableById("test")
	removeTable(o, table)
	if len(o.tables) != 0 {
		t.Error("Table not removed correctly.")
	}
}

func TestCheckTableLoaded(t *testing.T) {
	o := MakeOrchestrator()
	if o.IsTableLoaded("test") {
		t.Error("Table not correctly identified as not being loaded.")
	}
	o.GetTableById("test")
	if !o.IsTableLoaded("test") {
		t.Error("Table not correctly identified as being loaded.")
	}
}
