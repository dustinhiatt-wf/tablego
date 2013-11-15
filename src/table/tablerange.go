package table

import (
	"encoding/json"
	"strconv"
)

type tablerange struct {
	cells map[int]map[int]ICell
}

type valuerange struct {
	ISerializable
	Values map[string]map[string]string
}

func (vr *valuerange) Sum() string {
	sum := 0.0
	for row := range vr.Values {
		for column := range vr.Values[row] {
			value := vr.Values[row][column]
			i, ok := strconv.ParseInt(value, 10, 64)
			if ok == nil {
				sum = sum + float64(i)
				continue
			}
			f, ok := strconv.ParseFloat(value, 64)
			if ok == nil {
				sum = sum + f
			}
		}
	}
	return strconv.FormatFloat(sum, 'f', -1, 64)
}

func (vr *valuerange) ToBytes() []byte {
	res, err := json.Marshal(vr)
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

func MakeValueRange(cells map[int]map[int]ICell, cr *cellrange) *valuerange {
	vr := new(valuerange)
	vr.Values = make(map[string]map[string]string)
	for i := cr.StartRow; i < cr.StopRow; i++ {
		row, ok := cells[i]
		if !ok {
			continue
		}
		vr.Values[strconv.Itoa(i)] = make(map[string]string)
		for j := cr.StartColumn; j < cr.StopColumn; j++ {
			cell, ok := row[j]
			if ok {

				vr.Values[strconv.Itoa(i)][strconv.Itoa(j)] = cell.DisplayValue()
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
