package table

import (
	"testing"
	"time"
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
	if vr.Values["1"]["1"].CellDisplayValue != "test" {
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

func TestSumValueRange(t *testing.T) {
	vr := new(valuerange)
	vr.Values = make(map[string]map[string]*cellValue)
	vr.Values["1"] = make(map[string]*cellValue)
	vr.Values["2"] = make(map[string]*cellValue)
	vr.Values["1"]["4"] = makeCellValue("5.5", "5.5", time.Now())
	vr.Values["2"]["532"] = makeCellValue("7", "7", time.Now())
	vr.Values["2"]["0"] = makeCellValue("test", "test", time.Now())
	if vr.Sum() != "12.5" {
		t.Error("Valuerange not summing correctly.")
	}
}
