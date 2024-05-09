package main

import (
	"fmt"
	"net/http"
	"uno/internal"
)

func server(port string) {
	http.HandleFunc("/create", internal.CreateRoomHandler)
	http.HandleFunc("/join", internal.JoinRoomHandler)

	fmt.Printf("Server running on port %s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}
}
