package main

import (
	"fmt"
	"net/http"
	"uno/internal"
)

var players = []string{"p1", "p2"}

func server(port string) {
	game := internal.NewGame(players)
	go game.Network.BroadcastMessages()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		game.Network.EstablishConnections(w, r, game)
	})
	fmt.Printf("Server running on port %s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}
}
