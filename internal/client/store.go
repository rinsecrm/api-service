package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/rinsecrm/api-service/internal/canaryctx"
	pb "github.com/rinsecrm/api-service/proto"
)

type StoreClient struct {
	client pb.StoreServiceClient
	conn   *grpc.ClientConn
}

func NewStoreClient(address string) (*StoreClient, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(canaryctx.UnaryClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to store service: %w", err)
	}

	client := pb.NewStoreServiceClient(conn)

	return &StoreClient{
		client: client,
		conn:   conn,
	}, nil
}

func (s *StoreClient) Close() error {
	return s.conn.Close()
}

func (s *StoreClient) CreateItem(ctx context.Context, tenantID int64, name, description string, price float64, category pb.ItemCategory, sku string, inventoryCount int32, tags []string, createdBy string) (*pb.Item, error) {
	resp, err := s.client.CreateItem(ctx, &pb.CreateItemRequest{
		TenantId:       tenantID,
		Name:           name,
		Description:    description,
		Price:          price,
		Category:       category,
		Sku:            sku,
		InventoryCount: inventoryCount,
		Tags:           tags,
		CreatedBy:      createdBy,
	})
	if err != nil {
		return nil, err
	}
	return resp.Item, nil
}

func (s *StoreClient) GetItem(ctx context.Context, tenantID int64, id string) (*pb.Item, error) {
	resp, err := s.client.GetItem(ctx, &pb.GetItemRequest{
		TenantId: tenantID,
		Id:       id,
	})
	if err != nil {
		return nil, err
	}
	return resp.Item, nil
}

func (s *StoreClient) UpdateItem(ctx context.Context, tenantID int64, id, name, description string, price float64, category pb.ItemCategory, status pb.ItemStatus, sku string, inventoryCount int32, tags []string, updatedBy string) (*pb.Item, error) {
	resp, err := s.client.UpdateItem(ctx, &pb.UpdateItemRequest{
		TenantId:       tenantID,
		Id:             id,
		Name:           name,
		Description:    description,
		Price:          price,
		Category:       category,
		Status:         status,
		Sku:            sku,
		InventoryCount: inventoryCount,
		Tags:           tags,
		UpdatedBy:      updatedBy,
	})
	if err != nil {
		return nil, err
	}
	return resp.Item, nil
}

func (s *StoreClient) DeleteItem(ctx context.Context, tenantID int64, id string) (bool, error) {
	resp, err := s.client.DeleteItem(ctx, &pb.DeleteItemRequest{
		TenantId: tenantID,
		Id:       id,
	})
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}

func (s *StoreClient) ListItems(ctx context.Context, tenantID int64, category pb.ItemCategory, status pb.ItemStatus, searchQuery string, pageSize int32, pageToken string) ([]*pb.Item, string, int32, error) {
	resp, err := s.client.ListItems(ctx, &pb.ListItemsRequest{
		TenantId:    tenantID,
		Category:    category,
		Status:      status,
		SearchQuery: searchQuery,
		PageSize:    pageSize,
		PageToken:   pageToken,
	})
	if err != nil {
		return nil, "", 0, err
	}
	return resp.Items, resp.NextPageToken, resp.TotalCount, nil
}

func (s *StoreClient) UpdateInventory(ctx context.Context, tenantID int64, itemID string, quantityChange int32, reason, updatedBy string) (*pb.Item, int32, error) {
	resp, err := s.client.UpdateInventory(ctx, &pb.UpdateInventoryRequest{
		TenantId:       tenantID,
		ItemId:         itemID,
		QuantityChange: quantityChange,
		Reason:         reason,
		UpdatedBy:      updatedBy,
	})
	if err != nil {
		return nil, 0, err
	}
	return resp.Item, resp.PreviousCount, nil
}
