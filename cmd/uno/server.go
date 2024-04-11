package main

import (
	"fmt"
	"net/http"
	"uno/internal"
	"uno/internal/network"
)

var players = []string{"p1", "p2"}

func server(port string) {
	game := internal.NewGame(players)
	go network.BroadcastMessages(game)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		network.HandleConnections(w, r, game)
	})
	fmt.Printf("Server running on port %s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}
}
