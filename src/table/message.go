/**
 * Created with IntelliJ IDEA.
 * User: dustinhiatt
 * Date: 11/4/13
 * Time: 10:57 AM
 * To change this template use File | Settings | File Templates.
 */
package table

type valuemessage struct {
	operation 	string
	cell		*cell
}

type tablemessage struct {
	operation	string
}

func MakeValueChannel() chan *valuemessage {
	return make(chan *valuemessage)
}

func MakeTableChannel() chan *tablemessage {
	return make(chan *tablemessage, 1)
}
