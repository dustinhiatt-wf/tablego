package server

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"table"
	"node"
	"sync"
	"strconv"
)

// connection is an middleman between the websocket connection and the hub.
type connection struct {
        // The websocket connection.
        ws *websocket.Conn

        // Buffered channel of outbound messages.
        send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
        defer func() {
			c.ws.Close()
        }()
		msg := make(map[string]interface{})
        for {
			err := c.ws.ReadJSON(&msg)
			if err != nil {
				log.Println(err)
				break
			}
			switch msg["operation"] {
			case "register":
				cr := table.MakeRangeFromMap(msg)
				ch, vrCh := node.MakeMessageChannel(), node.MakeMessageChannel()

				var wg sync.WaitGroup
				wg.Add(1)
				go func () {
					wg.Done()
					for {
						select {
						case message := <- ch:
							response := make(map[string]interface{})
							response["operation"] = table.CellUpdated
							resp := make([]interface{}, 3)
							loc, _ := message.SourceCoordinates().(table.ITableCoordinates)
							resp[0] = loc.CellLocation().Row()
							resp[1] = loc.CellLocation().Column()
							cell := table.MakeCellFromBytes(message.Payload())
							resp[2] = cell.DisplayValue()
							response["values"] = resp
							c.write(response)
						}
					}
				}()
				wg.Wait()
				table.SubscribeToTableRange(cr, vrCh, ch)
				msg := <- vrCh
				vr := table.MakeValueRangeFromBytes(msg.Payload())
				resp := make(map[string]interface{})
				resp["operation"] = "registered"
				values := make([]interface{}, 0)
				for row, _ := range vr.Values {
					for column, _ := range vr.Values[row] {
						value := vr.Values[row][column]
						if value != ""{
							cell := make([]interface{}, 3)
							cell[0], _ = strconv.Atoi(row)
							cell[1], _ = strconv.Atoi(column)
							cell[2] = value
							values = append(values, cell)
						}
					}
				}
				resp["values"] = values
				c.write(resp)
			case table.EditCellValue:
				tableId := msg["table_id"].(string)
				row := int(msg["row"].(float64))
				column := int(msg["column"].(float64))
				value := msg["value"].(string)
				table.UpdateCellAtLocation(tableId, row, column, value)
			}
        }
}

// write writes a message with the given message type and payload.
func (c *connection) write(payload interface{}) error {
	err := c.ws.WriteJSON(payload)
	if err != nil {
		log.Println(err)
	}
	return err
}

// serverWs handles webocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}
	c.readPump()
}
