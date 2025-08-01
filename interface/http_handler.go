package interface

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"hotelsdatapipeline/config/domain"

	"github.com/gorilla/mux"
)

// HTTPHandler handles HTTP requests
type HTTPHandler struct {
	repository domain.HotelRepository
}

// APIResponse represents the standard API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Count   int         `json:"count,omitempty"`
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(repository domain.HotelRepository) *HTTPHandler {
	return &HTTPHandler{
		repository: repository,
	}
}

// HealthCheck handles health check requests
func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Success: true,
		Data:    "Hotel service is running",
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetHotelByID handles requests to get a hotel by ID
func (h *HTTPHandler) GetHotelByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hotelID := vars["id"]

	if hotelID == "" {
		response := APIResponse{
			Success: false,
			Error:   "Hotel ID is required",
		}
		h.writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	hotel, err := h.repository.GetHotelByID(hotelID)
	if err != nil {
		log.Printf("Failed to get hotel %s: %v", hotelID, err)
		response := APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Hotel not found: %s", hotelID),
		}
		h.writeJSONResponse(w, http.StatusNotFound, response)
		return
	}

	response := APIResponse{
		Success: true,
		Data:    hotel,
		Count:   1,
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetHotelsByDestination handles requests to get hotels by destination ID
func (h *HTTPHandler) GetHotelsByDestination(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	destinationIDStr := vars["id"]

	destinationID, err := strconv.Atoi(destinationIDStr)
	if err != nil {
		response := APIResponse{
			Success: false,
			Error:   "Invalid destination ID",
		}
		h.writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	hotels, err := h.repository.GetHotelsByDestinationID(destinationID)
	if err != nil {
		log.Printf("Failed to get hotels for destination %d: %v", destinationID, err)
		response := APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get hotels for destination %d", destinationID),
		}
		h.writeJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	response := APIResponse{
		Success: true,
		Data:    hotels,
		Count:   len(hotels),
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetHotelsByIDRange handles requests to get multiple hotels by IDs
func (h *HTTPHandler) GetHotelsByIDRange(w http.ResponseWriter, r *http.Request) {
	idsParam := r.URL.Query().Get("ids")
	if idsParam == "" {
		response := APIResponse{
			Success: false,
			Error:   "Hotel IDs parameter is required",
		}
		h.writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Parse and clean hotel IDs
	hotelIDs := strings.Split(idsParam, ",")
	var cleanHotelIDs []string
	
	for _, id := range hotelIDs {
		cleanID := strings.TrimSpace(id)
		if cleanID != "" {
			cleanHotelIDs = append(cleanHotelIDs, cleanID)
		}
	}

	if len(cleanHotelIDs) == 0 {
		response := APIResponse{
			Success: false,
			Error:   "No valid hotel IDs provided",
		}
		h.writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Limit to maximum 50 IDs
	if len(cleanHotelIDs) > 50 {
		response := APIResponse{
			Success: false,
			Error:   "Maximum 50 hotel IDs allowed per request",
		}
		h.writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	hotels, err := h.repository.GetHotelsByIDRange(cleanHotelIDs)
	if err != nil {
		log.Printf("Failed to get hotels by ID range: %v", err)
		response := APIResponse{
			Success: false,
			Error:   "Failed to get hotels",
		}
		h.writeJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	response := APIResponse{
		Success: true,
		Data:    hotels,
		Count:   len(hotels),
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// writeJSONResponse writes a JSON response with proper headers
func (h *HTTPHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
} 