package main

import (
	"log"

	"github.com/sant470/distlogs/internal/server"
)

func main() {
	srv := server.NewHTTPServer(":8081")
	log.Fatal(srv.ListenAndServe())
}
