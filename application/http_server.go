package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"hotelsdatapipeline/config/domain"
	"hotelsdatapipeline/httpinterface"
)

type HTTPServer struct {
	server *http.Server
	router *httpinterface.Router
}

func NewHTTPServer(host string, port int, repository domain.HotelRepository) *HTTPServer {
	router := httpinterface.NewRouter(repository)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &HTTPServer{
		server: server,
		router: router,
	}
}

func (hs *HTTPServer) Start() error {
	log.Printf("Starting HTTP server on %s", hs.server.Addr)

	if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func (hs *HTTPServer) Stop() error {
	log.Println("Stopping HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := hs.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	log.Println("HTTP server stopped gracefully")
	return nil
}

func (hs *HTTPServer) GetAddress() string {
	return hs.server.Addr
}
