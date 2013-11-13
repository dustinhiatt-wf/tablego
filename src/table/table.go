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
)

/*
type ITable interface {
	EditTableValue(row, column int, value string, ch chan *valuemessage)
	GetValueAt(row, column int, ch chan *valuemessage)
	GetRangeByRowAndColumn(startRow, stopRow, startColumn, stopColumn int, ch chan *valuemessage)
	GetRangeByCellRange(cr *cellrange, ch chan *valuemessage)
}

type table struct {
	id						string
	cells					map[int]map[int]*cell
	rows					map[int]chan *valuemessage
	columns					map[int]chan *valuemessage
	isinitialized			bool
	tablechannel			chan *valuemessage
	subscribers				*subscribers
	orchestrator			*orchestrator
	orchestratorChannel 	chan *valuemessage
}

/*
The idea here is that if you call into a channel to get a table message back
what you actually get back is something that resembles a table but is actually
just forwarding your commands and data to where the table actually lives

type tableobservable struct {
	channelToTable			chan *valuemessage
}

func MakeTableObservable(channelToTable chan *valuemessage) *tableobservable {
	to := new(tableobservable)
	to.channelToTable = channelToTable
	return to
}


/*
Listens for cell changes and notifies anyone listening to the table

func listenToCells(t *table, ch chan *valuemessage) {
	for {
		select {
		case message := <- ch:
			t.subscribers.notifySubscribers(message, false)
		}
	}
}

func getCell(table *table, row, column int) *cell {
	_, ok := table.cells[row]
	if ok {
		cell, ok := table.cells[row][column]
		if ok {
			return cell
		}
	}
	return nil
}

func tableHasCell(table *table, row, column int) bool {
	return getCell(table, row, column) != nil
}

func listenToOrchestrator(t *table) {
	for {
		select {
		case message := <- t.orchestratorChannel:
			switch message.operation {
			case "getTable":
				t.orchestratorChannel <- MakeValueMessage(GetTable, "", nil, nil, nil, t)
			case EditTableValue:

			}
		}
	}
}

func MakeTable(id string, orchestrator *orchestrator, ch chan *valuemessage) *table {
	table := new(table)
	table.id = id
	cells := make(map[int]map[int]*cell)
	table.cells = cells
	table.tablechannel = MakeValueChannel()
	table.subscribers = MakeSubscribers()
	table.orchestrator = orchestrator
	table.isinitialized = true
	table.orchestratorChannel = ch
	go func () {
		ch <- MakeValueMessage("tableOpened", "", nil, nil, nil, table)
	}()
	go listenToOrchestrator(table)
	return table
}

func (t *table) EditTableValue(row, column int, value string, ch chan *valuemessage) {
	row_item, ok := t.cells[row]
	if ok {
		cell, ok := row_item[column]
		if ok {
			cell.SetValue(value)
		} else {
			ch := MakeValueChannel()
			cell := MakeCell(row, column, value, t, ch)
			go listenToCells(t, ch)
			t.cells[row][column] = cell
		}
	} else {
		t.cells[row] = make(map[int]*cell)
		ch := MakeValueChannel()
		go listenToCells(t, ch)
		cell := MakeCell(row, column, value, t, ch)
		t.cells[row][column] = cell
	}
	go func () {
		ch <- MakeValueMessage(EditCell, "", t.cells[row][column], nil, nil, nil)
	}()
}

func (t *table) Subscribe(ch chan *valuemessage) {
	t.subscribers.append(ch)
}

func (t *table) GetValueAt(row, column int, ch chan *valuemessage) {
	cell := getCell(t, row, column)
	go func () {
		ch <- MakeValueMessage(GetValueAt, "", cell, nil, nil, nil)
	}()
}

func (t *table) GetRangeByRowAndColumn(startRow, stopRow, startColumn, stopColumn int, ch chan *valuemessage) {
	cr := &cellrange{startRow, stopRow + 1, startColumn, stopColumn + 1, t.id}
	t.GetRangeByCellRange(cr, ch)
}

func (t *table) GetRangeByCellRange(cr *cellrange, ch chan *valuemessage) {
	tr := MakeTableRange(t.cells, cr)
	go func () {
		ch <- MakeValueMessage(GetCellRange, "", nil, nil, tr, nil)
	}()
}
*/

type table struct {
	tableId			string
	orchestrator	*tableorchestratorchannel
	subscribers		*subscribers
	cells			map[int]map[int]*cellChannel
}

type cellChannel struct {
	channel				*tableCellChannel
	pendingRequests		map[string]*subscribers
	cellInitialized		bool
}

func (cc *cellChannel) sendNotification(operation string, message IMessage) {
	subs, ok := cc.pendingRequests[operation]
	if ok {
		subs.notifySubscribers(message, true)
	}
}

func (cc *cellChannel) subscribe(operation string, ch chan IMessage) {
	subs, ok := cc.pendingRequests[operation]
	if !ok {
		cc.pendingRequests[operation] = MakeSubscribers()
		subs = cc.pendingRequests[operation]
	}
	subs.append(ch)
}

func MakeCellChannel() *cellChannel {
	cc := new(cellChannel)
	cc.channel = MakeTableCellChannel()
	cc.pendingRequests = make(map[string]*subscribers)
	cc.cellInitialized = false
	return cc
}

type tableCellChannel struct {
	tableToCell		chan IMessage
	cellToTable		chan IMessage
}

func MakeTableCellChannel() *tableCellChannel {
	ch := new(tableCellChannel)
	ch.tableToCell = MakeMessageChannel()
	ch.cellToTable = MakeMessageChannel()
	return ch
}

func (t *table) getCellByPosition(row, column int, client chan IMessage) {
	_, ok := t.cells[row]
	if !ok {
		t.cells[row] = make(map[int]*cellChannel)
	}
	cc, ok := t.cells[row][column]
	if !ok {
		log.Println("CREATING CELL")
		go t.createCell(row, column, "",  client)
	} else if !cc.cellInitialized { //currently being loaded
		log.Println("WAITING FOR CELL INIT")
		cc.subscribe(CellOpened, client)
	} else { // table is loaded and ready
		log.Println("CELL FOUND")
		log.Println(cc)
		go func () {
			client <- MakeCommand(CellOpened, t.tableId, "", MakeCellLocation(row, column), nil, nil)
		}()
	}
}

func (t *table) createCell(row, column int, value string, client chan IMessage) {
	cc := MakeCellChannel()
	_, ok := t.cells[row]
	if !ok {
		t.cells[row] = make(map[int]*cellChannel)
	}

	t.cells[row][column] = cc
	t.cells[row][column].subscribe(CellOpened, client)
	go t.listenToCell(cc)
	MakeCell(row, column, value, cc)
}

func (t *table) sendToCell(cc *cellChannel, msg IMessage, ch chan IMessage) {
	cc.subscribe(msg.MessageId(), ch)
	cc.channel.tableToCell <- msg
}

func (t *table) listenToCell(cc *cellChannel) {
	for {
		select {
		case message := <- cc.channel.cellToTable:
			log.Println("TABLE GOT MESSAGE FROM CELL")
			if message.TargetTable() != "" && message.TargetTable() != t.tableId {
				log.Println("IN TABLE TO ORCHESTRATOR")
				go t.send(message, t.orchestrator.tableToOrchestrator)
			} else {
				switch message.Operation() {
				case CellOpened:
					cc.cellInitialized = true
					go cc.sendNotification(CellOpened, message)
				default:
					go cc.sendNotification(message.MessageId(), message)
				}
			}
		}
	}
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
			t.cells[row] = make(map[int]*cellChannel)
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

func (t *table) TableId() string {
	return t.tableId
}

func (t *table) send(msg IMessage, ch chan IMessage) {
	ch <- msg
}

func (t *table) forwardToChild(msg IMessage) {
	if msg.TargetCell() == nil {
		return
	}

	ch := MakeMessageChannel()
	t.getCellByPosition(msg.TargetCell().Row(), msg.TargetCell().Column(), ch)
	<- ch //to confirm it exists
	log.Println("FINAL FORWARD TO CELL")
	log.Println(msg.TargetCell().Row())
	log.Println(msg.TargetCell().Column())
	log.Println(t.cells[msg.TargetCell().Row()][msg.TargetCell().Column()].channel.tableToCell)
	t.cells[msg.TargetCell().Row()][msg.TargetCell().Column()].channel.tableToCell <- msg
}

func (t *table) Listen() {
	go func() {
		for {
			select {
			case message := <- t.orchestrator.orchestratorToTable:
				log.Println(message.TargetTable())
				log.Println(t.tableId)
				if message.TargetTable() != t.tableId {
					continue
				}
				if message.TargetCell() != nil {
					log.Println("FORWARDING TO CELL")
					go t.forwardToChild(message)
					continue
				}

				if message.GetType() == Response {
					continue
				}

				switch message.Operation() {
				case CloseTable:
					return
				case GetValueRange:
					go t.getValueRangeByCellRange(message, t.orchestrator.tableToOrchestrator)
				case EditCellValue:
					go t.editCellValue(message, t.orchestrator.tableToOrchestrator)
				default:
					log.Println("DEFAULT")
					go t.send(MakeResponse(message, nil), t.orchestrator.tableToOrchestrator)
				}
			}
		}
	}()
}

func MakeTable(tableId string, ch *tableorchestratorchannel) *table {
	t := new(table)
	t.tableId = tableId
	t.orchestrator = ch
	t.subscribers = MakeSubscribers()
	t.cells = make(map[int]map[int]*cellChannel)
	t.Listen()
	go t.send(MakeCommand(TableOpened, "", tableId, nil, nil, nil), t.orchestrator.tableToOrchestrator)
	return t
}
