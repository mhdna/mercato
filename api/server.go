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
	branchHub  *branchHub
	adminHub   *adminHub
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
		branchHub:  newBranchHub(),
		adminHub:   newAdminHub(),
	}

	server.setupRoutes()

	return server, nil
}

func (server *Server) setupRoutes() {
	router := gin.Default()
	router.Use(corsMiddleware())

	authRoutes := router.Group("/").Use(server.authMiddleware())

	authRoutes.POST("/inventories", server.createInventory)
	authRoutes.GET("/inventories/:id", server.getInventory)
	authRoutes.GET("/inventories/", server.listInventories)
	authRoutes.PUT("/inventories", server.updateInventory)
	authRoutes.DELETE("/inventories/:id", server.deleteInventory)
	authRoutes.POST("/products", server.createProduct)
	authRoutes.PUT("/products", server.updateProduct)
	authRoutes.GET("/products/:id", server.getProduct)
	authRoutes.GET("/products", server.listProducts)
	authRoutes.PUT("/product_variants", server.updateProductVariant)
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
	authRoutes.GET("/clients/:id/loyalty_total", server.getClientLoyaltyTotal)
	authRoutes.GET("/clients/", server.listClients)
	authRoutes.DELETE("/clients/:id", server.deleteClient)
	authRoutes.POST("/currencies", server.createCurrency)
	authRoutes.GET("/currencies/:code", server.getCurrency)
	authRoutes.GET("/currencies/", server.listCurrencies)
	authRoutes.PUT("/currencies", server.updateCurrency)
	authRoutes.DELETE("/currencies/:code", server.deleteCurrency)

	authRoutes.POST("/cashboxes", server.createCashbox)
	authRoutes.GET("/cashboxes/:id", server.getCashbox)
	authRoutes.GET("/cashboxes/", server.listCashboxes)
	authRoutes.PUT("/cashboxes", server.updateCashbox)
	authRoutes.POST("/shifts", server.createShift)
	authRoutes.POST("/shifts/:id/close", server.CloseShift)
	authRoutes.GET("/shifts", server.listShifts)
	authRoutes.GET("/shifts/:id", server.getShift)
	authRoutes.POST("/cashbox_accounts", server.createCashboxAccount)
	authRoutes.GET("/cashbox_accounts/", server.listCashboxAccounts)
	authRoutes.PUT("/cashbox_accounts", server.updateCashboxAccount)
	authRoutes.POST("/cashbox_accounts/balance", server.addCashboxAccountBalance)

	authRoutes.GET("/pos_settings", server.posSettings)

	authRoutes.POST("/suppliers", server.createSupplier)
	authRoutes.GET("/suppliers/:id", server.getSupplier)
	authRoutes.GET("/suppliers", server.listSuppliers)

	authRoutes.POST("/colors", server.createColor)
	authRoutes.GET("/colors", server.listColors)
	authRoutes.POST("/sizes", server.createSize)
	authRoutes.GET("/sizes", server.listSizes)

	authRoutes.POST("/sales_invoices", server.createSalesInvoice)
	authRoutes.GET("/sales_invoices/:id", server.getSalesInvoice)
	authRoutes.GET("/sales_invoices", server.listSalesInvoices)

	authRoutes.POST("/return_invoices", server.createReturnInvoice)
	authRoutes.GET("/return_invoices/:id", server.getReturnInvoice)
	authRoutes.GET("/return_invoices", server.listReturnInvoices)

	authRoutes.POST("/invoice_types", server.createInvoiceType)
	authRoutes.GET("/invoice_types/:id", server.getInvoiceType)
	authRoutes.GET("/invoice_types", server.listInvoiceTypes)
	authRoutes.PUT("/invoice_types", server.updateInvoiceType)

	authRoutes.POST("/price_lists", server.createPriceList)
	authRoutes.GET("/price_lists/:id", server.getPriceList)
	authRoutes.GET("/price_lists", server.listPriceLists)
	authRoutes.POST("/price_lists/items", server.createPriceListItem)
	authRoutes.GET("/price_lists/:id/items", server.listPriceListItems)
	authRoutes.DELETE("/price_lists/:id/items/:product_id", server.deletePriceListItem)

	authRoutes.POST("/discount_lists", server.createDiscountList)
	authRoutes.GET("/discount_lists/:id", server.getDiscountList)
	authRoutes.GET("/discount_lists", server.listDiscountLists)
	authRoutes.PUT("/discount_lists", server.updateDiscountList)
	authRoutes.POST("/discount_lists/items", server.createDiscountListItem)
	authRoutes.GET("/discount_lists/:id/items", server.listDiscountListItems)
	authRoutes.DELETE("/discount_lists/:id/items/:product_id", server.deleteDiscountListItem)

	router.POST("/users/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)
	router.POST("/users", server.createUser)
	authRoutes.GET("/users/:id", server.getUser)
	authRoutes.GET("/users", server.listUsers)
	authRoutes.PUT("/users", server.updateUser)
	authRoutes.DELETE("/users/:id", server.deleteUser)

	authRoutes.POST("/entries", server.createEntry)
	authRoutes.GET("/entries/:id", server.getEntry)
	authRoutes.GET("/entries", server.listEntries)

	authRoutes.POST("/expense_categories", server.createExpenseCategory)
	authRoutes.GET("/expense_categories/:id", server.getExpenseCategory)
	authRoutes.GET("/expense_categories", server.listExpenseCategories)
	authRoutes.PUT("/expense_categories", server.updateExpenseCategory)
	authRoutes.DELETE("/expense_categories/:id", server.deleteExpenseCategory)

	authRoutes.POST("/expenses", server.createExpense)
	authRoutes.GET("/expenses/:id", server.getExpense)
	authRoutes.GET("/expenses", server.listExpenses)

	authRoutes.POST("/recurring_expenses", server.createRecurringExpense)
	authRoutes.GET("/recurring_expenses", server.listRecurringExpenses)
	authRoutes.PUT("/recurring_expenses/active", server.setRecurringExpenseActive)

	authRoutes.POST("/loan_categories", server.createLoanCategory)
	authRoutes.GET("/loan_categories/:id", server.getLoanCategory)
	authRoutes.GET("/loan_categories", server.listLoanCategories)
	authRoutes.PUT("/loan_categories", server.updateLoanCategory)
	authRoutes.DELETE("/loan_categories/:id", server.deleteLoanCategory)

	authRoutes.POST("/loans", server.createLoan)
	authRoutes.GET("/loans/:id", server.getLoan)
	authRoutes.GET("/loans", server.listLoans)

	authRoutes.POST("/loan_payments", server.createLoanPayment)
	authRoutes.GET("/loan_payments", server.listLoanPayments)
	authRoutes.DELETE("/loan_payments/:id", server.deleteLoanPayment)

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

	authRoutes.POST("/branches", server.createBranch)
	authRoutes.GET("/branches", server.listBranches)
	authRoutes.POST("/branches/:id/activate", server.setBranchActive(true))
	authRoutes.POST("/branches/:id/deactivate", server.setBranchActive(false))
	authRoutes.POST("/branches/:id/rotate_key", server.rotateBranchKey)

	authRoutes.GET("/branch_expenses", server.listBranchExpenses)
	authRoutes.GET("/branch_invoices", server.listBranchInvoices)
	authRoutes.GET("/branch_invoices/:id/items", server.listBranchInvoiceItems)
	authRoutes.GET("/branch_invoices/daily_income", server.dailyIncome)
	authRoutes.GET("/branches/:id/settings", server.getBranchSettings)
	authRoutes.POST("/branches/:id/commands", server.createBranchCommand)
	authRoutes.GET("/branches/:id/commands", server.listBranchCommands)

	authRoutes.GET("/branches/:id/targets", server.listBranchTargets)
	authRoutes.POST("/branches/:id/targets", server.createBranchTarget)
	authRoutes.PUT("/branch_targets/:id", server.updateBranchTarget)
	authRoutes.DELETE("/branch_targets/:id", server.deleteBranchTarget)

	authRoutes.GET("/branches/:id/salespersons", server.listSalespersons)
	authRoutes.POST("/branches/:id/salespersons", server.createSalesperson)
	authRoutes.PUT("/salespersons/:id", server.updateSalesperson)
	authRoutes.PUT("/salespersons/:id/active", server.setSalespersonActive)

	// adminWS can't sit under authRoutes: authMiddleware only reads the
	// Authorization header, and a browser WebSocket handshake can't set
	// custom headers. Auth happens inside adminWS itself via a query-param
	// token instead.
	router.GET("/admin/ws", server.adminWS)

	// branchRoutes is authenticated with a per-branch API key rather than a
	// PASETO user token — see branchAuthMiddleware. Kept under its own path
	// prefix so branch-facing request/response shapes (branch_id, client_ref
	// idempotency, etc.) never collide with the human-operator endpoints above.
	branchRoutes := router.Group("/branch").Use(server.branchAuthMiddleware())
	branchRoutes.GET("/health", server.branchHealth)
	branchRoutes.POST("/sales_invoices", server.createBranchSalesInvoice)
	branchRoutes.POST("/return_invoices", server.createBranchReturnInvoice)
	branchRoutes.POST("/expenses", server.createBranchExpense)
	branchRoutes.POST("/loans", server.createBranchLoan)
	branchRoutes.GET("/sync/changes", server.branchSyncChanges)
	branchRoutes.GET("/ws", server.branchWS)
	branchRoutes.PUT("/settings", server.putBranchSettings)
	branchRoutes.POST("/clients", server.putBranchClient)
	branchRoutes.GET("/commands/pending", server.branchPendingCommands)
	branchRoutes.POST("/commands/:id/ack", server.ackBranchCommand)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
