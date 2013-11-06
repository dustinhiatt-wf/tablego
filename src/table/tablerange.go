/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/4/13
 * Time: 9:00 PM
 * To change this template use File | Settings | File Templates.
 */
package table

type tablerange struct {
	cells 	 	map[int]map[int]*cell
}

type valuerange struct {
	values 		map[int]map[int]string
}

func MakeTableRange(cells map[int]map[int]*cell, cr *cellrange) *tablerange {
	tb := new(tablerange)
	tb.cells = make(map[int]map[int]*cell)
	for i := cr.startRow; i < cr.stopRow; i++ {
		row, ok := cells[i]
		if !ok {
			continue
		}
		tb.cells[i] = make(map[int]*cell)
		for j := cr.startColumn; j < cr.stopColumn; j++ {
			cell, ok := row[j]
			if ok {
				tb.cells[i][j] = cell
			}
		}
	}
	return tb
}

func MakeValueRange(cells map[int]map[int]*cell, cr *cellrange) *valuerange {
	vr := new(valuerange)
	vr.values = make(map[int]map[int]string)
	for i := cr.startRow; i < cr.stopRow; i++ {
		row, ok := cells[i]
		if !ok {
			continue
		}
		vr.values[i] = make(map[int]string)
		for j := cr.startColumn; j < cr.stopColumn; j++ {
			cell, ok := row[j]
			if ok {
				vr.values[i][j] = cell.DisplayValue
			}
		}
	}
	return vr
}
