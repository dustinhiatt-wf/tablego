/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/6/13
 * Time: 10:26 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"testing"
)

func TestMakeRangeWithTableId(t *testing.T) {
	cr := MakeRange("test:A1:C3")
	if cr.tableId != "test" {
		t.Error("Table Id not set correctly.")
	} else if cr.startRow != 0 {
		t.Error("Start row not set correclty.")
	} else if cr.stopRow != 3 {
		t.Error("Stop row not set correctly.")
	} else if cr.startColumn != 0 {
		t.Error("Start column not set correctly.")
	} else if cr.stopColumn != 3 {
		t.Error("Stop column not set correctly.")
	}
}

func TestMakeRangeWithoutTableId(t *testing.T) {
	cr := MakeRange("A1:C3")
	if cr.tableId != "" {
		t.Error("Table Id not set correctly.")
	} else if cr.startRow != 0 {
		t.Error("Start row not set correclty.")
	} else if cr.stopRow != 3 {
		t.Error("Stop row not set correctly.")
	} else if cr.startColumn != 0 {
		t.Error("Start column not set correctly.")
	} else if cr.stopColumn != 3 {
		t.Error("Stop column not set correctly.")
	}
}
