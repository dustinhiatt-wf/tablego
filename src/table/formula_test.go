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
	if cr.TableId != "test" {
		t.Error("Table Id not set correctly.")
	} else if cr.StartRow != 0 {
		t.Error("Start row not set correclty.")
	} else if cr.StopRow != 3 {
		t.Error("Stop row not set correctly.")
	} else if cr.StartColumn != 0 {
		t.Error("Start column not set correctly.")
	} else if cr.StopColumn != 3 {
		t.Error("Stop column not set correctly.")
	}
}

func TestMakeRangeWithoutTableId(t *testing.T) {
	cr := MakeRange("A1:C3")
	if cr.TableId != "" {
		t.Error("Table Id not set correctly.")
	} else if cr.StartRow != 0 {
		t.Error("Start row not set correclty.")
	} else if cr.StopRow != 3 {
		t.Error("Stop row not set correctly.")
	} else if cr.StartColumn != 0 {
		t.Error("Start column not set correctly.")
	} else if cr.StopColumn != 3 {
		t.Error("Stop column not set correctly.")
	}
}
