package gapi

import (
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/mhdna/kashi/pb"
)

func convertInventoryType(t db.InventoryType) pb.InventoryType {
	switch t {
	case db.InventoryTypeWarehouse:
		return pb.InventoryType_INVENTORY_TYPE_WAREHOUSE
	case db.InventoryTypeStore:
		return pb.InventoryType_INVENTORY_TYPE_STORE
	default:
		return pb.InventoryType_INVENTORY_TYPE_UNSPECIFIED
	}
}

func convertInventory(inventory db.Inventory) *pb.Inventory {
	return &pb.Inventory{
		Name: inventory.Name,
		Code: inventory.Code,
		Type: convertInventoryType(inventory.Type),
	}
}
