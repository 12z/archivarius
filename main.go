package main

import (
	"net/http"

	"github.com/12z/archivarius/server"
)

func main() {
	srv := http.Server{}
	server := server.NewServer(&srv)
	server.Serve()
}
