/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/11/13
 * Time: 8:48 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"encoding/json"
	//"log"
)

/*
This represents all the commands that can be sent to a table structure.  This
is designed to go to and from JSON
*/
type tablecommand struct {
	Value string
}

func (tc *tablecommand) ToBytes() []byte {
	b, err := json.Marshal(tc)
	if err != nil {
		return nil
	}
	return b
}

func MakeTableCommand(value string) *tablecommand {
	tc := new(tablecommand)
	tc.Value = value
	return tc
}

func MakeTableCommandFromJson(bytes []byte) *tablecommand {
	var m tablecommand
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return nil
	}
	return &m
}

type tableresponse struct {
}
