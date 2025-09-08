package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/your-org/api-service/internal/canaryctx"
	"github.com/your-org/api-service/internal/client"
)

type Server struct {
	storeClient *client.StoreClient
}

func NewServer(storeClient *client.StoreClient) *Server {
	return &Server{
		storeClient: storeClient,
	}
}

type ItemRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Price          float64  `json:"price"`
	Category       string   `json:"category,omitempty"`
	SKU            string   `json:"sku,omitempty"`
	InventoryCount int32    `json:"inventory_count,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

type ItemUpdateRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Price          float64  `json:"price"`
	Category       string   `json:"category,omitempty"`
	Status         string   `json:"status,omitempty"`
	SKU            string   `json:"sku,omitempty"`
	InventoryCount int32    `json:"inventory_count,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

type ItemResponse struct {
	ID             string   `json:"id"`
	TenantID       int64    `json:"tenant_id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Price          float64  `json:"price"`
	Category       string   `json:"category"`
	Status         string   `json:"status"`
	SKU            string   `json:"sku"`
	InventoryCount int32    `json:"inventory_count"`
	Tags           []string `json:"tags"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	CreatedBy      string   `json:"created_by"`
	UpdatedBy      string   `json:"updated_by"`
}

type InventoryUpdateRequest struct {
	QuantityChange int32  `json:"quantity_change"`
	Reason         string `json:"reason"`
	UpdatedBy      string `json:"updated_by"`
}

type ListResponse struct {
	Items         []ItemResponse `json:"items"`
	NextPageToken string         `json:"next_page_token,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *Server) CreateItem(w http.ResponseWriter, r *http.Request) {
	// Log canary context for observability
	if canary, ok := canaryctx.FromContext(r.Context()); ok {
		log.Printf("CreateItem called with canary PR: %s", canary)
	}

	var req ItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	userID := getUserFromRequest(r)
	category := stringToCategory(req.Category)

	item, err := s.storeClient.CreateItem(
		r.Context(),
		tenantID,
		req.Name,
		req.Description,
		req.Price,
		category,
		req.SKU,
		req.InventoryCount,
		req.Tags,
		userID,
	)
	if err != nil {
		log.Printf("Error creating item: %v", err)
		writeErrorResponse(w, "Failed to create item", http.StatusInternalServerError)
		return
	}

	response := protoItemToResponse(item)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) GetItem(w http.ResponseWriter, r *http.Request) {
	if canary, ok := canaryctx.FromContext(r.Context()); ok {
		log.Printf("GetItem called with canary PR: %s", canary)
	}

	vars := mux.Vars(r)
	id := vars["id"]
	tenantID := getTenantIDFromRequest(r)

	item, err := s.storeClient.GetItem(r.Context(), tenantID, id)
	if err != nil {
		log.Printf("Error getting item: %v", err)
		writeErrorResponse(w, "Item not found", http.StatusNotFound)
		return
	}

	response := protoItemToResponse(item)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) UpdateItem(w http.ResponseWriter, r *http.Request) {
	if canary, ok := canaryctx.FromContext(r.Context()); ok {
		log.Printf("UpdateItem called with canary PR: %s", canary)
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var req ItemUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	userID := getUserFromRequest(r)
	category := stringToCategory(req.Category)
	status := stringToStatus(req.Status)

	item, err := s.storeClient.UpdateItem(
		r.Context(),
		tenantID,
		id,
		req.Name,
		req.Description,
		req.Price,
		category,
		status,
		req.SKU,
		req.InventoryCount,
		req.Tags,
		userID,
	)
	if err != nil {
		log.Printf("Error updating item: %v", err)
		writeErrorResponse(w, "Failed to update item", http.StatusInternalServerError)
		return
	}

	response := protoItemToResponse(item)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) DeleteItem(w http.ResponseWriter, r *http.Request) {
	if canary, ok := canaryctx.FromContext(r.Context()); ok {
		log.Printf("DeleteItem called with canary PR: %s", canary)
	}

	vars := mux.Vars(r)
	id := vars["id"]
	tenantID := getTenantIDFromRequest(r)

	success, err := s.storeClient.DeleteItem(r.Context(), tenantID, id)
	if err != nil {
		log.Printf("Error deleting item: %v", err)
		writeErrorResponse(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	if !success {
		writeErrorResponse(w, "Item not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ListItems(w http.ResponseWriter, r *http.Request) {
	if canary, ok := canaryctx.FromContext(r.Context()); ok {
		log.Printf("ListItems called with canary PR: %s", canary)
	}

	tenantID := getTenantIDFromRequest(r)

	pageSize := int32(10) // default
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil {
			pageSize = int32(parsed)
		}
	}

	pageToken := r.URL.Query().Get("page_token")
	categoryFilter := stringToCategory(r.URL.Query().Get("category"))
	statusFilter := stringToStatus(r.URL.Query().Get("status"))
	searchQuery := r.URL.Query().Get("search")

	items, nextPageToken, totalCount, err := s.storeClient.ListItems(r.Context(), tenantID, categoryFilter, statusFilter, searchQuery, pageSize, pageToken)
	if err != nil {
		log.Printf("Error listing items: %v", err)
		writeErrorResponse(w, "Failed to list items", http.StatusInternalServerError)
		return
	}

	var responseItems []ItemResponse
	for _, item := range items {
		responseItems = append(responseItems, protoItemToResponse(item))
	}

	response := ListResponse{
		Items:         responseItems,
		NextPageToken: nextPageToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	if canary, ok := canaryctx.FromContext(r.Context()); ok {
		log.Printf("UpdateInventory called with canary PR: %s", canary)
	}

	vars := mux.Vars(r)
	itemID := vars["id"]

	var req InventoryUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.UpdatedBy == "" {
		req.UpdatedBy = getUserFromRequest(r)
	}

	tenantID := getTenantIDFromRequest(r)

	item, previousCount, err := s.storeClient.UpdateInventory(
		r.Context(),
		tenantID,
		itemID,
		req.QuantityChange,
		req.Reason,
		req.UpdatedBy,
	)
	if err != nil {
		log.Printf("Error updating inventory: %v", err)
		writeErrorResponse(w, "Failed to update inventory", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"item":           protoItemToResponse(item),
		"previous_count": previousCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
