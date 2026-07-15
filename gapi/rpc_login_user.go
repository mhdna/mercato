package gapi

import (
	"context"

	"github.com/lib/pq"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/mhdna/kashi/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateInvetory(ctx context.Context, req *pb.CreateInventoryRequest) (*pb.CreateInventoryResponse, error) {
	arg := db.CreateInventoryParams{
		Name: req.GetName(),
		// FIXME: see if the one below works
		Type: db.InventoryType(req.GetType()),
		Code: req.GetCode(),
	}
	inventory, err := server.store.CreateInventory(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "inventory already exists: %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create inventory: %s", err)
	}
	rsp := &pb.CreateInventoryResponse{
		Inventory: convertInventory(inventory),
	}
	return rsp, nil
}
