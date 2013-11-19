package main

import (
        "log"
        "net/http"
		"server"
)

func main() {
		http.Handle("/", http.FileServer(http.Dir("./server/static")))
        http.HandleFunc("/ws", server.ServeWs)
        err := http.ListenAndServe(":8123", nil)
        if err != nil {
                log.Fatal("ListenAndServe: ", err)
        }
}
