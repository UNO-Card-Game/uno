package main

import "os"

func main() {
	port := os.Getenv("PORT")
	server(port)
}
