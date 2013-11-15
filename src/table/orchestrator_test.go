/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/5/13
 * Time: 3:54 PM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	"node"
	"testing"
)

func TestCreateOrchestrator(t *testing.T) {
	o := MakeOrchestrator()
	if o.INode.Parent() != nil {
		t.Error("Orchestrator created with parent.")
	}
	coords := MakeCoordinates("", nil)

	if !o.INode.Coordinates().Equal(coords) {
		t.Error("Orchestrator created with wrong coordinates.")
	}
	if o.INode.ParentCoordinates() != nil {
		t.Error("Orchestrator created with parent coordinates.")
	}
}

func TestCreateChildTable(t *testing.T) {
	o := MakeOrchestrator()
	ch := node.MakeMessageChannel()
	o.INode.GetOrCreateChild(ch, MakeCoordinates("test", nil))
	<-ch //table initialized
	if len(o.children) != 1 {
		t.Error("Child table not created.")
	}
}

func TestOrchestratorReturnsCachedTable(t *testing.T) {
	o := MakeOrchestrator()
	ch := node.MakeMessageChannel()
	o.INode.GetOrCreateChild(ch, MakeCoordinates("test", nil))
	<-ch
	o.INode.GetOrCreateChild(ch, MakeCoordinates("test", nil))
	<-ch
	if len(o.children) != 1 {
		t.Error("Child cached more than once")
	}
}

func TestOrchestratorNotifiesOnTableCreation(t *testing.T) {
	o := MakeOrchestrator()
	ch := node.MakeMessageChannel()
	chTwo := node.MakeMessageChannel()
	o.INode.GetOrCreateChild(ch, MakeCoordinates("test", nil))
	o.INode.GetOrCreateChild(chTwo, MakeCoordinates("test", nil))
	<-ch
	<-chTwo
	if len(o.children) != 1 {
		t.Error("Child tables not created correctly.")
	}
}

func TestOrchestratorRoutesToChild(t *testing.T) {
	o := MakeOrchestrator()
	ch := node.MakeMessageChannel()
	o.INode.GetOrCreateChild(ch, MakeCoordinates("test", nil))
	<-ch
	tblCh := node.MakeIChild()
	o.children["test2"] = tblCh
	o.children["test"].Channel().ChildToParent() <- node.MakeCommand("test", MakeCoordinates("test2", nil), MakeCoordinates("test", nil), nil)
	msg := <-tblCh.Channel().ParentToChild()
	if msg.Operation() != "test" {
		t.Error("Orchestrator did not forward message correctly.")
	}
}
