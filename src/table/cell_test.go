/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/4/13
 * Time: 11:57 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"testing"
	//"log"
)

func TestAsInt(t *testing.T) {
	v := MakeCell(1, 1, "7", nil, nil)
	value, success := v.AsInt()
	if value != 7 || success != true {
		t.Error("Value data does not equal int64 7")
	}
}

func TestAsFloat(t *testing.T) {
	v := MakeCell(1, 1, "7.1", nil, nil)
	value, success := v.AsFloat()
	if value != 7.1 || success != true {
		t.Error("Value data does not equal float64 7.1")
	}
}

func TestAsFloatConversionFailure(t *testing.T) {
	v := MakeCell(1, 1, "a", nil, nil)
	_, success := v.AsFloat()
	if success != false {
		t.Error("Conversion of 'a' to float succeeded.")
	}
}

func TestAsIntConversionFailure(t *testing.T) {
	v := MakeCell(1, 1, "a", nil, nil)
	_, success := v.AsInt()
	if success != false {
		t.Error("Convert of 'a' to int succeeded.")
	}
}

func TestCreateValue(t *testing.T) {
	v := MakeCell(1, 1, "", nil, nil)
	ticks := v.timestamp
	if ticks == 0 {
		t.Error("Ticks are zero")
	}
}

func TestMakeCell(t *testing.T) {
	cell := MakeCell(1, 1, "test", nil, nil)
	if cell.row != 1 {
		t.Error("Row not set correctly.")
	} else if cell.column != 1 {
		t.Error("Column not set correctly.")
	} else if cell.value != "test" {
		t.Error("Value not set correctly.")
	}
}

func TestUpdateValue(t *testing.T) {
	cell := MakeCell(1, 1, "test", nil, nil)
	cell.SetValue("test2")
	if cell.value != "test2" {
		t.Error("Cell didn't update for channel message.")
	}
}

func TestExcelLetterToIndex(t *testing.T) {
	index := getNumberFromAlpha("a")
	if index != 0 {
		t.Error("Wrong number returned for alpha.")
	}
}

func TestExcelLetterToIndexMultipleCharacters(t *testing.T) {
	index := getNumberFromAlpha("aa")
	if index != 26 {
		t.Error("Wrong number returned for alpha.")
	}
}

func TestParseFormula(t *testing.T) {
	funcName := "=test(foo, bar)"
	funcParts := parseFormula(funcName)
	if funcParts[0] != "test" || funcParts[1] != "foo, bar" {
		t.Error("Function not parsed correctly.")
	}
}

func TestParseValueForFormula(t *testing.T) {
	funcName := "=test(foo, bar)"
	cell := MakeCell(1, 1, funcName, nil, nil)
	parseValueForFormula(cell)
	if cell.isFormula == false {
		t.Error("IsFormula not marked correctly.")
	}
}

func TestGetPartsFromAlphaNumeric(t *testing.T) {
	alphaNumeric := "A1"
	parts := getStringPartsFromAlphaNumeric(alphaNumeric)
	if parts[0] != "A" || parts[1] != "1" {
		t.Error("Alphanumeric not parsed correctly.")
	}
}

func TestGetNumberFromAlphaNumeric(t *testing.T) {
	alphaNumeric := "1"
	parts := getStringPartsFromAlphaNumeric(alphaNumeric)
	if len(parts) != 1 || parts[0] != "1" {
		t.Error("Alphanumeric not parsed correctly.")
	}
}

func TestGetLetterFromAlphaNumeric(t *testing.T) {
	alphaNumeric := "A"
	parts := getStringPartsFromAlphaNumeric(alphaNumeric)
	if len(parts) != 1 || parts[0] != "A" {
		t.Error("Alphanumeric not parsed correctly.")
	}
}

func TestParseAlphaNumericParts(t *testing.T) {
	parts := []string{"A", "1"}
	row, column := parseAlphaNumericParts(parts)
	if row != 0 || column != 0 {
		t.Error("Alphanumeric parts not parsed correctly.")
	}
}

func TestParseAlphaNumericPartsNumberOnly(t *testing.T) {
	parts := []string{"1"}
	row, column := parseAlphaNumericParts(parts)
	if row != 0 || column != -1 {
		t.Error("Alphanumeric parts not parsed correctly.")
	}
}

func TestParseAlphaNumericPartsLetterOnly(t *testing.T) {
	parts := []string{"A"}
	row, column := parseAlphaNumericParts(parts)
	if row != -1 || column != 0 {
		t.Error("Alphanumeric parts not parsed correctly.")
	}
}

func TestCreateRange(t *testing.T) {
	cr := MakeRange("A1:A1")
	if cr.startRow != 0 {
		t.Error("Start row not set correctly.")
	} else if cr.stopRow != 1 {
		t.Error("Stop row not set correctly.")
	} else if cr.startColumn != 0 {
		t.Error("Start column not set correctly.")
	} else if cr.stopColumn != 1 {
		t.Error("Stop column not set correctly.")
	}
}

func TestSubscribeToCell(t *testing.T) {
	cell := MakeCell(1, 1, "1", nil, nil)
	ch := MakeValueChannel()
	cell.Subscribe(ch)
	cell.SetValue("2")
	message := <- ch
	if message.cell.DisplayValue != "2" {
		t.Error("Subscribe to cell not functioning properly.")
	}
}
