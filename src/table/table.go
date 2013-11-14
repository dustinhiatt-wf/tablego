/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/4/13
 * Time: 9:17 AM
 * To change this template use File | Settings | File Templates.
 */
package table


import (
	"strconv"
	"log"
	"node"
)

type table struct {
	node.INode
	node.INodeFactory
	node.ICommunicationHandler
	children 								map[int]map[int]node.IChild
}

func (t *table) onMessageFromParent(msg IMessage) {
	if msg.GetType() == node.Response {

	} else if msg.GetType() == node.Command {

	} else {
		//TODO: log error
	}
}

func (t *table) onMessageFromChild(msg IMessage) {
	if msg.GetType() == node.Response {

	} else if msg.GetType() == node.Command {

	} else {
		//TODO: log error
	}
}

func (t *table) GetChild(coords ICoordinates) IChild {
	original, _ := coords.(ITableCoordinates)
	_, ok := t.children[original.CellLocation().Row()]
	if !ok {
		return nil
	}
	child, ok := t.children[original.CellLocation().Row()][original.CellLocation().Column()]
	if !ok {
		return nil
	}
	return child
}

func (t *table) makeChildNode(parentChannel IChild, childCoordinates ICoordinates) INode {
	child := MakeTable(parentChannel, childCoordinates, t.INode.Coordinates())
	loc, _ := childCoordinates.(ITableCoordinates)
	o.children[loc.TableId()] = parentChannel
	return child
}

func (t *table) getValueRangeByCellRange(cmd IMessage, ch chan IMessage) {
	cr := MakeRangeFromBytes(cmd.Payload())
	if cr == nil {
		go t.send(MakeError(cmd, "Error parsing cell range."), ch)
	}
	go func (){
		vr := new(valuerange)
		vr.Values = make(map[string]map[string]string)
		listeners := make([]chan IMessage, 0)
		for i := cr.StartRow; i < cr.StopRow; i++ {
			_, ok := t.cells[i]
			if !ok {
				continue
			}
			vr.Values[strconv.Itoa(i)] = make(map[string]string)
			for j := cr.StartColumn; j < cr.StopColumn; j++ {
				cc, ok := t.cells[i][j]
				if ok {
					ch := MakeMessageChannel()
					go t.sendToCell(cc, MakeCommand(GetCellValue, t.tableId, t.tableId, MakeCellLocation(i, j), nil, nil), ch)
					listeners = append(listeners, ch)
					vr.Values[strconv.Itoa(i)][strconv.Itoa(j)] = ""
				}
			}
		}
		for _, ch := range listeners {
			message := <- ch
			cell := MakeCellFromBytes(message.Payload())
			vr.Values[strconv.Itoa(message.SourceCell().Row())][strconv.Itoa(message.SourceCell().Column())] = cell.DisplayValue()
		}

		ch <- MakeResponse(cmd, vr.ToBytes())
	}()
}

func (t *table) editCellValue(cmd IMessage, ch chan IMessage) {
	row := cmd.TargetCell().Row()
	column := cmd.TargetCell().Column()

	go func() {
		_, ok := t.cells[row]
		if !ok {
			t.cells[row] = make(map[int]node.IChild)
		}

		cc, ok := t.cells[row][column]
		if !ok {
			ch := MakeMessageChannel()
			go t.createCell(row, column, "", ch)
			<- ch // cell is created
			cc = t.cells[row][column]
		}

		go t.sendToCell(cc, cmd, ch)
	}()
}



func MakeTable(parentChannel IChannel, coordinates, parentCoordinates ICoordinates) *table {
	t := new(table)
	t.cells = make(map[int]map[int]node.IChild)
	// this is where we need to load and parse
	t.INode = node.MakeNode(parentChannel, coordinates, parentCoordinates, t, t)
	return t
}
