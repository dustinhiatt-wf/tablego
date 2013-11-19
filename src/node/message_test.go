package node

import (
	"testing"
	"bytes"
)

func TestMakeCommand(t *testing.T) {
	payload := make([]byte, 0)
	sourceCell := makeCoordinates("test")
	targetCell := makeCoordinates("test")
	command := MakeCommand("test", targetCell, sourceCell, nil)
	if command.Operation() != "test" {
		t.Error("Operation not set correctly.")
	} else if command.TargetCoordinates() != targetCell {
		t.Error("Target coordinates not set correctly.")
	} else if command.SourceCoordinates() != sourceCell {
		t.Error("Source coordinates not set correctly.")
	} else if !bytes.Equal(payload, command.Payload()) {
		t.Error("Payload not set correctly.")
	}
}

func TestMakeResult(t *testing.T) {
	payload := make([]byte, 0)
	sourceCell := makeCoordinates("test")
	targetCell := makeCoordinates("test2")
	command := MakeCommand("test", targetCell, sourceCell, nil)
	response := MakeResponse(command, payload)
	if response.Operation() != command.Operation() {
		t.Error("Response operation not set correctly.")
	} else if response.TargetCoordinates() != command.SourceCoordinates() {
		t.Error("Response target table not set correctly.")
	} else if response.SourceCoordinates() != command.TargetCoordinates() {
		t.Error("Response target cell not set correctly.")
	} else if response.MessageId() != command.MessageId() {
		t.Error("Response message id not set correctly.")
	} else if !bytes.Equal(payload, response.Payload()) {
		t.Error("Response payload not set correctly.")
	}
}

/*
func TestMessageMarshal(t *testing.T) {
	msg := makeMessage("test", makeCoordinates("test"), makeCoordinates("test"), make([]byte, 0))
	bytes := msg.ToBytes()
	log.Println(bytes)
	res := MakeMessageFromBytes(bytes)
	log.Println(res.TargetCoordinates())
	if !reflect.DeepEqual(msg, res) {
		t.Error("IMessage couldn't be marshaled and unmarshaled.")
	}
}*/
