package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"practicum-webinars/devops/internal/server"
)

const (
	address        = "localhost:8080"
	shudownTimeout = 5 * time.Second
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := http.Server{
		Addr:    address,
		Handler: server.NewRouter(),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		fmt.Println("Start listen monitor server on " + address)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Println("HTTP server ListenAndServe:", err)
		}
	}()

	termSignal := make(chan os.Signal, 1)
	signal.Notify(termSignal, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-termSignal
	fmt.Println("Shutting down... reason:", sig.String())

	server.StopClients()

	ctx, cancel = context.WithTimeout(ctx, shudownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("shutdown error:", err)
		return
	}
	fmt.Println("gracefully stopped")
}
