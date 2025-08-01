package interface

import (
	"log"
	"net/http"
	"time"

	"hotelsdatapipeline/config/domain"

	"github.com/gorilla/mux"
)

// Router handles HTTP routing and middleware
type Router struct {
	router     *mux.Router
	repository domain.HotelRepository
	handler    *HTTPHandler
}

// NewRouter creates a new router with middleware
func NewRouter(repository domain.HotelRepository) *Router {
	router := mux.NewRouter()
	handler := NewHTTPHandler(repository)

	r := &Router{
		router:     router,
		repository: repository,
		handler:    handler,
	}

	r.setupMiddleware()
	r.setupRoutes()

	return r
}

// setupMiddleware configures logging and CORS middleware
func (r *Router) setupMiddleware() {
	// Logging middleware
	r.router.Use(r.loggingMiddleware)
	
	// CORS middleware
	r.router.Use(r.corsMiddleware)
}

// setupRoutes configures all API routes
func (r *Router) setupRoutes() {
	api := r.router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", r.handler.HealthCheck).Methods("GET")

	// Hotel endpoints
	api.HandleFunc("/hotels/{id}", r.handler.GetHotelByID).Methods("GET")
	api.HandleFunc("/hotels/destination/{id}", r.handler.GetHotelsByDestination).Methods("GET")
	api.HandleFunc("/hotels/range", r.handler.GetHotelsByIDRange).Methods("GET")

	// Handle preflight OPTIONS requests
	api.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// loggingMiddleware logs request details
func (r *Router) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		
		next.ServeHTTP(w, req)
		
		log.Printf(
			"%s %s %s %v",
			req.Method,
			req.RequestURI,
			req.RemoteAddr,
			time.Since(start),
		)
	})
}

// corsMiddleware handles CORS headers
func (r *Router) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, req)
	})
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
} 