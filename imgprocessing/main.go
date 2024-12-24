package main

import (
	"fmt"
	processor "github.com/mayusGomez/imgprocessing/pipelineprocessor"
	"github.com/mayusGomez/imgprocessing/server"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	serv := server.Server{Port: ":8080"}
	serv.Start()
	time.Sleep(2 * time.Second)

	processor.Start(50)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Program is running. Press Ctrl+C to exit.")
	sig := <-sigChan

	fmt.Printf("Received signal: %s. Shutting down...\n", sig)
}
