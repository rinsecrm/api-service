package server

import (
	"net/http"
	"strconv"

	pb "github.com/rinsecrm/api-service/proto"
)

// Helper functions for converting between HTTP request/response and proto types

func stringToCategory(category string) pb.ItemCategory {
	switch category {
	case "electronics":
		return pb.ItemCategory_ITEM_CATEGORY_ELECTRONICS
	case "clothing":
		return pb.ItemCategory_ITEM_CATEGORY_CLOTHING
	case "books":
		return pb.ItemCategory_ITEM_CATEGORY_BOOKS
	case "home":
		return pb.ItemCategory_ITEM_CATEGORY_HOME
	case "sports":
		return pb.ItemCategory_ITEM_CATEGORY_SPORTS
	default:
		return pb.ItemCategory_ITEM_CATEGORY_UNSPECIFIED
	}
}

func categoryToString(category pb.ItemCategory) string {
	switch category {
	case pb.ItemCategory_ITEM_CATEGORY_ELECTRONICS:
		return "electronics"
	case pb.ItemCategory_ITEM_CATEGORY_CLOTHING:
		return "clothing"
	case pb.ItemCategory_ITEM_CATEGORY_BOOKS:
		return "books"
	case pb.ItemCategory_ITEM_CATEGORY_HOME:
		return "home"
	case pb.ItemCategory_ITEM_CATEGORY_SPORTS:
		return "sports"
	default:
		return "unspecified"
	}
}

func stringToStatus(status string) pb.ItemStatus {
	switch status {
	case "active":
		return pb.ItemStatus_ITEM_STATUS_ACTIVE
	case "inactive":
		return pb.ItemStatus_ITEM_STATUS_INACTIVE
	case "out_of_stock":
		return pb.ItemStatus_ITEM_STATUS_OUT_OF_STOCK
	case "discontinued":
		return pb.ItemStatus_ITEM_STATUS_DISCONTINUED
	default:
		return pb.ItemStatus_ITEM_STATUS_UNSPECIFIED
	}
}

func statusToString(status pb.ItemStatus) string {
	switch status {
	case pb.ItemStatus_ITEM_STATUS_ACTIVE:
		return "active"
	case pb.ItemStatus_ITEM_STATUS_INACTIVE:
		return "inactive"
	case pb.ItemStatus_ITEM_STATUS_OUT_OF_STOCK:
		return "out_of_stock"
	case pb.ItemStatus_ITEM_STATUS_DISCONTINUED:
		return "discontinued"
	default:
		return "unspecified"
	}
}

func protoItemToResponse(item *pb.Item) ItemResponse {
	var tags []string
	if item.Tags != nil {
		tags = item.Tags
	} else {
		tags = []string{}
	}

	return ItemResponse{
		ID:             item.Id,
		TenantID:       item.TenantId,
		Name:           item.Name,
		Description:    item.Description,
		Price:          item.Price,
		Category:       categoryToString(item.Category),
		Status:         statusToString(item.Status),
		SKU:            item.Sku,
		InventoryCount: item.InventoryCount,
		Tags:           tags,
		CreatedAt:      item.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      item.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z"),
		CreatedBy:      item.CreatedBy,
		UpdatedBy:      item.UpdatedBy,
	}
}

// getTenantIDFromRequest extracts tenant ID from request context or headers
// In a real implementation, this would be extracted from JWT token or similar auth mechanism
func getTenantIDFromRequest(r *http.Request) int64 {
	// For demo purposes, use a header. In production, extract from authenticated context
	if tenantHeader := r.Header.Get("X-Tenant-ID"); tenantHeader != "" {
		// Parse tenant ID from header
		if tenantID, err := strconv.ParseInt(tenantHeader, 10, 64); err == nil {
			return tenantID
		}
	}
	// Default tenant for demo
	return 1
}

// getUserFromRequest extracts user ID from request context or headers
// In a real implementation, this would be extracted from JWT token or similar auth mechanism
func getUserFromRequest(r *http.Request) string {
	// For demo purposes, use a header. In production, extract from authenticated context
	if userHeader := r.Header.Get("X-User-ID"); userHeader != "" {
		return userHeader
	}
	// Default user for demo
	return "system"
}
