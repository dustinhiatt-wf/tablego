/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/11/13
 * Time: 10:50 AM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"bytes"
	"testing"
)

func TestMakeCommand(t *testing.T) {
	payload := make([]byte, 0)
	sourceCell := MakeCellLocation(1, 1)
	targetCell := MakeCellLocation(2, 2)
	command := MakeCommand("test", "targetTable", "sourceTable", targetCell, sourceCell, payload)
	if command.Operation() != "test" {
		t.Error("Operation not set correctly.")
	} else if command.TargetCell() != targetCell {
		t.Error("Target cell not set correctly.")
	} else if command.TargetTable() != "targetTable" {
		t.Error("Target table not set correctly.")
	} else if command.SourceCell() != sourceCell {
		t.Error("Source cell not set correctly.")
	} else if command.SourceTable() != "sourceTable" {
		t.Error("Source table not set correctly.")
	} else if !bytes.Equal(payload, command.Payload()) {
		t.Error("Payload not set correctly.")
	}
}

func TestMakeResult(t *testing.T) {
	payload := make([]byte, 0)
	sourceCell := MakeCellLocation(1, 1)
	targetCell := MakeCellLocation(2, 2)
	command := MakeCommand("test", "targetTable", "sourceTable", targetCell, sourceCell, payload)
	response := MakeResponse(command, payload)
	if response.Operation() != command.Operation() {
		t.Error("Response operation not set correctly.")
	} else if response.TargetTable() != command.SourceTable() {
		t.Error("Response target table not set correctly.")
	} else if response.TargetCell() != command.SourceCell() {
		t.Error("Response target cell not set correctly.")
	} else if response.SourceCell() != command.TargetCell() {
		t.Error("Response source cell not set correctly.")
	} else if response.SourceTable() != command.TargetTable() {
		t.Error("Response source table not set correctly.")
	} else if response.MessageId() != command.MessageId() {
		t.Error("Response message id not set correctly.")
	} else if !bytes.Equal(payload, response.Payload()) {
		t.Error("Response payload not set correctly.")
	}
}
