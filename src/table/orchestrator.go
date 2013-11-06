/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/5/13
 * Time: 3:36 PM
 * To change this template use File | Settings | File Templates.
 */
package table

var master = MakeOrchestrator()

type orchestrator struct {
	tables		[]*table
}

func appendTable(o *orchestrator, t *table) {
	if o.IsTableLoaded(t.id) {
		return
	}
	o.tables = append(o.tables, t)
}

func removeTable(o *orchestrator, t *table) {
	i := -1
	for index, table := range o.tables {
		if table.id == t.id {
			i = index
		}
	}
	if i == -1 {
		return
	}
	o.tables = append(o.tables[:i], o.tables[i+1:]...)
}

func (o *orchestrator) IsTableLoaded(id string) bool {
	for _, table := range o.tables {
		if table.id == id {
			return true
		}
	}
	return false
}

func (o *orchestrator) GetTableById(id string) *table {
	var t *table = nil
	for _, table := range o.tables {
		if table.id == id {
			t = table
			break
		}
	}
	if t == nil {
		t = MakeTable(id, o)
		appendTable(o, t)
	}
	return t
}

func MakeOrchestrator() *orchestrator {
	o := new(orchestrator)
	o.tables = make([]*table, 0)
	return o
}
