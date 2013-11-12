/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/5/13
 * Time: 3:36 PM
 * To change this template use File | Settings | File Templates.
 */
package table

import (
//	"log"
)

var master = MakeOrchestrator()

type tableorchestratorchannel struct {
	orchestratorToTable		chan IMessage
	tableToOrchestrator		chan IMessage
}

func MakeTableOrchestratorChannel() *tableorchestratorchannel {
	ch := new(tableorchestratorchannel)
	ch.orchestratorToTable = make(chan IMessage)
	ch.tableToOrchestrator = make(chan IMessage)
	return ch
}

type tablechannel struct {
	channel				*tableorchestratorchannel
	pendingRequests		map[string]*subscribers
	tableInitialized	bool
}

type orchestrator struct {
	tables		map[string]*tablechannel
}

func (tc *tablechannel) sendNotification(operation string, message IMessage) {
	subs, ok := tc.pendingRequests[operation]
	if ok {
		subs.notifySubscribers(message, true)
	}
}

func (tc *tablechannel) subscribe(operation string, ch chan IMessage) {
	subs, ok := tc.pendingRequests[operation]
	if !ok {
		tc.pendingRequests[operation] = MakeSubscribers()
		subs = tc.pendingRequests[operation]
	}
	subs.append(ch)
}

func MakeTableChannel() *tablechannel {
	tc := new(tablechannel)
	tc.channel = MakeTableOrchestratorChannel()
	tc.pendingRequests = make(map[string]*subscribers)
	tc.tableInitialized = false // not really needed but being verbose
	return tc
}



func removeTable(o *orchestrator, tableId string) {
	tc, ok := o.tables[tableId]
	if ok {
		tc.channel.orchestratorToTable <- MakeCommand(CloseTable, tableId, "", nil, nil, nil)
	}
	delete(o.tables, tableId)
}

func (o *orchestrator) IsTableLoaded(id string) bool {
	_, ok := o.tables[id]
	return ok
}

func (o *orchestrator) sendCommand(cmd ICommand, ch chan IMessage) {
	go func () {
		if cmd.TargetTable() == "" {
			return
		}
		tableCh := MakeMessageChannel()
		var tc *tablechannel
		go func () {
			for {
				select {
				case message := <- tableCh:
					if message.Operation() == TableOpened {
						tc = o.tables[cmd.TargetTable()]
						internalSendCommand(cmd, tc, ch)
						return
					}

				}
			}
		}()
		o.GetTableById(cmd.TargetTable(), tableCh) //make sure table is ready to go
	}()
}

func listenToTable(tc *tablechannel) {
	for {
		select {
		case message := <- tc.channel.tableToOrchestrator:
			switch message.Operation() {
			case TableOpened:
				tc.tableInitialized = true
				tc.sendNotification(TableOpened, message)
			default:
				tc.sendNotification(message.MessageId(), message)

			}
		}
	}
}

func createTable(ch *tableorchestratorchannel, id string) {
	MakeTable(id, ch)
}

func internalSendCommand(cmd ICommand, tc *tablechannel, ch chan IMessage) {
	tc.subscribe(cmd.MessageId(), ch)
	tc.channel.orchestratorToTable <- cmd
}

func SendCommand(cmd ICommand, ch chan IMessage) {
	master.sendCommand(cmd, ch)
}

func (o *orchestrator) GetTableById(id string, client chan IMessage) {
	tc, ok := o.tables[id]
	if !ok {
		o.tables[id] = MakeTableChannel()
		o.tables[id].subscribe("tableOpened", client)
		go listenToTable(o.tables[id])
		go createTable(o.tables[id].channel, id)
	} else if !tc.tableInitialized { //currently being loaded
		tc.subscribe(TableOpened, client)
	} else { // table is loaded and ready
		go func () {
			client <- MakeCommand(TableOpened, id, "", nil, nil, nil)
		}()
	}
}

func MakeOrchestrator() *orchestrator {
	o := new(orchestrator)
	o.tables = make(map[string]*tablechannel)
	return o
}
