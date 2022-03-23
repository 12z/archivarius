package main

import "github.com/12z/archivarius/server"

func main() {
	server := server.NewServer()
	server.Serve()
}
