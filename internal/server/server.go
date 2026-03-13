package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run(ctx context.Context, r http.Handler) {
	server := &http.Server{
		Addr: ":8080",
		Handler: r,
	}

	go func() {
		log.Println("Server started on http://localhost:8080")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("The server caught a sad one:", err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		syscall.SIGINT,
        syscall.SIGTERM,
	)

	defer signal.Stop(quit)

	<-quit
	log.Println("Shutting down server...")

    if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced shutdown:", err)
	}

	log.Println("Server stopped")
}