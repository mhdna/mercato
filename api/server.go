package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/mhdna/kashi/token"
	"github.com/mhdna/kashi/util"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %s", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	server.setupRoutes()

	return server, nil
}

func (server *Server) setupRoutes() {
	router := gin.Default()

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/inventories", server.createInventory)
	authRoutes.GET("/inventories/:id", server.getInventory)
	authRoutes.GET("/inventories/", server.listInventories)
	authRoutes.POST("/products", server.createProduct)
	authRoutes.GET("/products/:id", server.getProduct)
	authRoutes.GET("/products", server.listProducts)
	authRoutes.GET("/barcodes", server.listBarcodes)
	authRoutes.DELETE("/products/:id", server.deleteProduct)
	authRoutes.POST("/attributes/batch", server.createAttributeValues)
	authRoutes.POST("/attributes", server.createAttributeValue)
	authRoutes.PUT("/attributes", server.updateAttributeValue)
	authRoutes.GET("/attributes/", server.listAttributeValues)
	// TODO: add updateAsset
	authRoutes.POST("/assets", server.createAsset)
	authRoutes.DELETE("/assets/:id", server.deleteAsset)
	authRoutes.GET("/assets/:id", server.getAsset)
	authRoutes.GET("/assets/", server.listAssets)
	// TODO: add getAssetType
	authRoutes.POST("/asset_types", server.createAssetType)
	authRoutes.DELETE("/asset_types/:id", server.deleteAssetType)
	authRoutes.POST("/clients", server.createClient)
	authRoutes.PUT("/clients", server.updateClient)
	authRoutes.GET("/clients/:id", server.getClient)
	authRoutes.GET("/clients/", server.listClients)
	authRoutes.POST("/currencies", server.createCurrency)
	authRoutes.GET("/currencies/:id", server.getCurrency)
	authRoutes.GET("/currencies/", server.listCurrencies)

	authRoutes.POST("/cashboxes", server.createCashbox)
	authRoutes.GET("/cashboxes/:id", server.getCashbox)
	authRoutes.GET("/cashboxes/", server.listCashboxes)
	authRoutes.POST("/shifts", server.createShift)
	authRoutes.POST("/shifts/:id/close", server.CloseShift)
	authRoutes.GET("/shifts", server.listShifts)
	authRoutes.GET("/shifts/:id", server.getShift)
	authRoutes.POST("/cashbox_accounts", server.createCashboxAccount)
	authRoutes.GET("/cashbox_accounts/", server.listCashboxAccounts)
	authRoutes.PUT("/cashbox_accounts", server.updateCashboxAccount)
	authRoutes.POST("/cashbox_accounts/balance", server.addCashboxAccountBalance)

	authRoutes.POST("/suppliers", server.createSupplier)
	authRoutes.GET("/suppliers/:id", server.getSupplier)
	authRoutes.GET("/suppliers", server.listSuppliers)

	authRoutes.POST("/sales_invoices", server.createSalesInvoice)
	authRoutes.GET("/sales_invoices/:id", server.getSalesInvoice)
	authRoutes.GET("/sales_invoices", server.listSalesInvoices)

	authRoutes.POST("/return_invoices", server.createReturnInvoice)
	authRoutes.GET("/return_invoices/:id", server.getReturnInvoice)
	authRoutes.GET("/return_invoices", server.listReturnInvoices)

	authRoutes.POST("/price_lists", server.createPriceList)
	authRoutes.GET("/price_lists/:id", server.getPriceList)
	authRoutes.GET("/price_lists", server.listPriceLists)
	authRoutes.POST("/price_lists/items", server.createPriceListItem)
	authRoutes.GET("/price_lists/:id/items", server.listPriceListItems)
	authRoutes.DELETE("/price_lists/:id/items/:product_id", server.deletePriceListItem)

	authRoutes.POST("/discount_lists", server.createDiscountList)
	authRoutes.GET("/discount_lists/:id", server.getDiscountList)
	authRoutes.GET("/discount_lists", server.listDiscountLists)
	authRoutes.POST("/discount_lists/items", server.createDiscountListItem)
	authRoutes.GET("/discount_lists/:id/items", server.listDiscountListItems)
	authRoutes.DELETE("/discount_lists/:id/items/:product_id", server.deleteDiscountListItem)

	authRoutes.POST("/users", server.createUser)
	authRoutes.GET("/users/:id", server.getUser)
	authRoutes.GET("/users", server.listUsers)
	authRoutes.PUT("/users", server.updateUser)
	authRoutes.DELETE("/users/:id", server.deleteUser)

	authRoutes.POST("/entries", server.createEntry)
	authRoutes.GET("/entries/:id", server.getEntry)
	authRoutes.GET("/entries", server.listEntries)

	authRoutes.POST("/expenses", server.createExpense)
	authRoutes.GET("/expenses/:id", server.getExpense)
	authRoutes.GET("/expenses", server.listExpenses)

	authRoutes.POST("/coupons", server.createCoupon)
	authRoutes.GET("/coupons/:code", server.getCoupon)
	authRoutes.GET("/coupons", server.listCoupons)
	authRoutes.PUT("/coupons/:code/deactivate", server.deactivateCoupon)

	authRoutes.POST("/transfers", server.createTransfer)
	authRoutes.GET("/transfers/:id", server.getTransfer)
	authRoutes.GET("/transfers", server.listTransfers)
	authRoutes.PUT("/transfers", server.updateTransfer)
	authRoutes.POST("/transfer_items", server.createTransferItem)
	authRoutes.GET("/transfer_items/:transfer_id", server.listTransferItems)

	authRoutes.POST("/purchases", server.createPurchase)
	authRoutes.GET("/purchases/:id", server.getPurchase)
	authRoutes.GET("/purchases", server.listPurchases)
	authRoutes.POST("/purchases/items", server.addPurchaseItem)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) *gin.H {
	return &gin.H{"error": err.Error()}
}
