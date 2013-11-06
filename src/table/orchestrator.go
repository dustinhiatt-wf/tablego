/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/5/13
 * Time: 3:36 PM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
	//"log"
)

var master = MakeOrchestrator()

type tablechannel struct {
	channel				chan *valuemessage
	pendingRequests		map[string]*subscribers
	tableInitialized	bool
}

func (tc *tablechannel) sendNotification(operation string, message *valuemessage) {
	subs, ok := tc.pendingRequests[operation]
	if ok {
		subs.notifySubscribers(message, true)
	}
}

func (tc *tablechannel) subscribe(operation string, ch chan *valuemessage) {
	subs, ok := tc.pendingRequests[operation]
	if !ok {
		tc.pendingRequests[operation] = MakeSubscribers()
		subs = tc.pendingRequests[operation]
	}
	subs.append(ch)
}

func MakeTableChannel(ch chan *valuemessage) *tablechannel {
	tc := new(tablechannel)
	tc.channel = ch
	tc.pendingRequests = make(map[string]*subscribers)
	tc.tableInitialized = false // not really needed but being verbose
	return tc
}

type orchestrator struct {
	tables		map[string]*tablechannel
}

func removeTable(o *orchestrator, tableId string) {
	tc, ok := o.tables[tableId]
	if ok {
		tc.channel <- MakeValueMessage(TableClosed, "", nil, nil, nil, nil)
	}
	delete(o.tables, tableId)
}

func (o *orchestrator) IsTableLoaded(id string) bool {
	_, ok := o.tables[id]
	return ok
}

func listenToTable(ch *tablechannel) {
	for {
		select {
		case message := <- ch.channel:
			switch message.operation{
			case "tableClosed":
				ch.sendNotification(TableClosed, message)
				return
			case "tableOpened":
				ch.tableInitialized = true
				ch.sendNotification(TableOpened, message)
			case GetTable:
				if message.table != nil {
					ch.sendNotification(GetTable, message)
				}
			}
		}
	}
}

func createTable(o *orchestrator, id string) {
	MakeTable(id, o, o.tables[id].channel)
}

func (o *orchestrator) GetTableById(id string, client chan *valuemessage) {
	tc, ok := o.tables[id]
	if !ok {
		server := MakeValueChannel()
		o.tables[id] = MakeTableChannel(server)
		o.tables[id].subscribe("tableOpened", client)
		go listenToTable(o.tables[id])
		go createTable(o, id)
	} else if !tc.tableInitialized { //currently being loaded
		tc.subscribe(TableOpened, client)
	} else { // table is loaded and ready
		tc.subscribe(GetTable, client)
		go func () {
			tc.channel <- MakeValueMessage(GetTable, "", nil, nil, nil, nil)
		}()
	}
}

func MakeOrchestrator() *orchestrator {
	o := new(orchestrator)
	o.tables = make(map[string]*tablechannel)
	return o
}
