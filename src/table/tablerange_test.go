/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/5/13
 * Time: 8:06 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"testing"
)

func TestMakeTableRange(t *testing.T) {
	table := MakeTable("test", nil)
	table.EditTableValue(1, 1, "test")
	tr := MakeTableRange(table.cells, MakeRange("A1:C3"))
	if tr.cells[1][1] != table.cells[1][1] {
		t.Error("Table range not made correctly.")
	}
}

