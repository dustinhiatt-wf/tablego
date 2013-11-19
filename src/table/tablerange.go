package table

import (
	"encoding/json"
	"strconv"
	"sync"
	"strings"
)

type tablerange struct {
	ISerializable
	cells map[int]map[int]ICell
}

func (tr *tablerange) ToBytes() []byte {
	res, err := json.Marshal(tr)
	if err != nil {
		return nil
	}
	return res
}

func MakeTableRange(cells map[int]map[int]ICell, cr *cellrange) *tablerange {
	tb := new(tablerange)
	tb.cells = make(map[int]map[int]ICell)
	for i := cr.StartRow; i < cr.StopRow; i++ {
		row, ok := cells[i]
		if !ok {
			continue
		}
		tb.cells[i] = make(map[int]ICell)
		for j := cr.StartColumn; j < cr.StopColumn; j++ {
			cell, ok := row[j]
			if ok {
				tb.cells[i][j] = cell
			}
		}
	}
	return tb
}

func MakeTableRangeFromBytes(bytes []byte) *tablerange {
	var m tablerange
	json.Unmarshal(bytes, &m)
	return &m
}

type valuerange struct {
	ISerializable
	Values 	map[string]map[string]*cellValue
	mutex	sync.Mutex
}

func (vr *valuerange) update(row, column int, cv *cellValue) {
	strRow := strconv.Itoa(row)
	strCol := strconv.Itoa(column)
	vr.mutex.Lock()
	_, ok := vr.Values[strRow]
	if ok {
		_, ok = vr.Values[strRow]
		if ok {
			vr.Values[strRow][strCol] = cv
		}
	}
	vr.mutex.Unlock()
}

func (vr *valuerange) Sum() string {
	sum := 0.0
	vr.mutex.Lock()
	for row := range vr.Values {
		for column := range vr.Values[row] {
			value := vr.Values[row][column]
			i, ok := strconv.ParseInt(value.CellDisplayValue, 10, 64)
			if ok == nil {
				sum = sum + float64(i)
				continue
			}
			f, ok := strconv.ParseFloat(value.CellDisplayValue, 64)
			if ok == nil {
				sum = sum + f
			}
		}
	}
	vr.mutex.Unlock()
	return strconv.FormatFloat(sum, 'f', -1, 64)
}

func (vr *valuerange) Vlookup(value string, index int, cellRange *cellrange) string {
	value = strings.TrimSpace(value)
	lookupColumn := strconv.Itoa(cellRange.StartColumn)
	for i := cellRange.StartRow; i < cellRange.StopRow; i++ {
		lookupRow := strconv.Itoa(i)
		_, ok := vr.Values[lookupRow]
		if ok {
			_, ok = vr.Values[lookupRow][lookupColumn]
			if ok {
				lookupValue := vr.Values[lookupRow][lookupColumn]
				if lookupValue.CellDisplayValue == value {
					indexColumn := strconv.Itoa(cellRange.StartColumn + index)
					_, ok = vr.Values[lookupRow][indexColumn]
					if ok {
						return vr.Values[lookupRow][indexColumn].CellDisplayValue
					} else {
						return ""
					}
				}
			}
		}
	}
	return ""
}

func (vr *valuerange) ToBytes() []byte {
	res, err := json.Marshal(vr)
	if err != nil {
		return nil
	}

	return res
}

func MakeValueRange(cells map[int]map[int]ICell, cr *cellrange) *valuerange {
	vr := new(valuerange)
	vr.Values = make(map[string]map[string]*cellValue)
	for i := cr.StartRow; i < cr.StopRow; i++ {
		row, ok := cells[i]
		if !ok {
			continue
		}
		vr.Values[strconv.Itoa(i)] = make(map[string]*cellValue)
		for j := cr.StartColumn; j < cr.StopColumn; j++ {
			cell, ok := row[j]
			if ok {

				vr.Values[strconv.Itoa(i)][strconv.Itoa(j)] = cell.GetCellValue()
			}
		}
	}
	return vr
}

func MakeValueRangeFromBytes(bytes []byte) *valuerange {
	var m valuerange
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return nil
	}
	return &m
}

func ConvertStringKeyedMapToIntKeys(stringKeyed map[string]map[string]ICell) map[int]map[int]ICell {
	intKeyed := make(map[int]map[int]ICell)
	for row := range stringKeyed {
		for column := range stringKeyed[row] {
			intRow, _ := strconv.Atoi(row)
			_, ok := intKeyed[intRow]
			if !ok {
				intKeyed[intRow] = make(map[int]ICell)
			}
			intColumn, _ := strconv.Atoi(column)
			intKeyed[intRow][intColumn] = stringKeyed[row][column]
		}
	}
	return intKeyed
}
