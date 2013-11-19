/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/11/13
 * Time: 9:05 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"encoding/json"
	"testing"
)

func TestTableCommandToJson(t *testing.T) {
	tc := MakeTableCommand("value")
	var f interface{}
	js := tc.ToBytes()
	json.Unmarshal(js, &f)
	m := f.(map[string]interface{})
	if m["Value"] != "value" {
		t.Error("Table value not set correctly.")
	}
}

func TestJsonToTableCommand(t *testing.T) {
	tc := MakeTableCommand("value")
	testTc := MakeTableCommandFromJson(tc.ToBytes())
	if tc.Value != testTc.Value {
		t.Error("Table value not recovered correctly.")
	}
}
