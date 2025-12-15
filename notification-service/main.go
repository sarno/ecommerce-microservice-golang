package main

import (
	"log"
	"net/http"
	"notification-service/cmd"
)


func main() {
	// Start pprof server in a goroutine
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Execute the Cobra command (which will start your main application server)
	cmd.Execute()
}
