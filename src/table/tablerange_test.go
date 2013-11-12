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
	tbl := make(map[int]map[int]ICell)
	tbl[1] = make(map[int]ICell)
	tbl[1][1] = new(cell)
	tr := MakeTableRange(tbl, MakeRange("A1:C3"))
	if tr.cells[1][1] != tbl[1][1] {
		t.Error("Table range not made correctly.")
	}
}

func TestMakeValueRange(t *testing.T) {
	tbl := make(map[int]map[int]ICell)
	tbl[1] = make(map[int]ICell)
	cell := new(cell)
	cell.CellDisplayValue = "test"
	tbl[1][1] = cell
	vr := MakeValueRange(tbl, MakeRange("A1:C3"))
	if vr.Values["1"]["1"] != "test" {
		t.Error("Value range was not made correctly.")
	}
}

func TestTurnStringKeyedToIntKeyed(t *testing.T) {
	tbl := make(map[string]map[string]ICell)
	tbl["1"] = make(map[string]ICell)
	cell := new(cell)
	tbl["1"]["1"] = cell
	rng := ConvertStringKeyedMapToIntKeys(tbl)
	if tbl["1"]["1"] != rng[1][1] {
		t.Error("String map not converted to int map.")
	}
}
