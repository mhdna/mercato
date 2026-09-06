package api

import (
	"fmt"
	"log"

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

// BranchHub exposes the server's live-push hub so background schedulers
// (e.g. RunBranchTargetSeriesScheduler) that run outside a request's
// lifecycle can notify connected branches too, using the same connections
// branch API handlers push through.
func (server *Server) BranchHub() *branchHub {
	return server.branchHub
}

func (server *Server) setupRoutes() {
	// Persist Gin access and recovery output through the same bounded writer
	// configured for the standard Go logger while keeping Gin's middleware.
	gin.DefaultWriter = log.Writer()
	gin.DefaultErrorWriter = log.Writer()
	router := gin.Default()
	router.Use(corsMiddleware())

	authRoutes := router.Group("/").Use(server.authMiddleware())

	authRoutes.POST("/inventories", server.createInventory)
	authRoutes.GET("/inventories/:id", server.getInventory)
	authRoutes.GET("/inventories/", server.listInventories)
	authRoutes.PUT("/inventories", server.updateInventory)
	authRoutes.DELETE("/inventories/:id", server.deleteInventory)
	authRoutes.POST("/inventories/bulk_delete", server.bulkDeleteInventories)
	authRoutes.GET("/inventories/:id/stock", server.listInventoryStock)
	authRoutes.POST("/inventories/:id/adjustments", server.createStockAdjustment)
	authRoutes.GET("/stock_movements", server.listStockMovements)
	authRoutes.POST("/products", server.createProduct)
	authRoutes.PUT("/products", server.updateProduct)
	authRoutes.GET("/products/:id", server.getProduct)
	authRoutes.GET("/products", server.listProducts)
	authRoutes.PUT("/product_variants", server.updateProductVariant)
	authRoutes.POST("/product_variants", server.createProductVariant)
	authRoutes.DELETE("/product_variants/:id", server.deleteProductVariant)
	authRoutes.GET("/barcodes", server.listBarcodes)
	authRoutes.POST("/barcodes/assign", server.assignBarcodes)
	authRoutes.POST("/barcodes/print", server.printBarcodes)
	authRoutes.DELETE("/products/:id", server.deleteProduct)
	authRoutes.POST("/attributes/batch", server.createAttributeValues)
	authRoutes.POST("/attributes", server.createAttributeValue)
	authRoutes.PUT("/attributes", server.updateAttributeValue)
	authRoutes.GET("/attributes/", server.listAttributeValues)
	authRoutes.GET("/attribute_types", server.listAttributes)
	authRoutes.GET("/attribute_values", server.listAllAttributeValues)
	authRoutes.DELETE("/attributes/:id", server.deleteAttributeValue)
	authRoutes.POST("/attributes/bulk_delete", server.bulkDeleteAttributeValues)
	authRoutes.POST("/assets", server.createAsset)
	authRoutes.PUT("/assets", server.updateAsset)
	authRoutes.DELETE("/assets/:id", server.deleteAsset)
	authRoutes.POST("/assets/bulk_delete", server.bulkDeleteAssets)
	authRoutes.GET("/assets/:id", server.getAsset)
	authRoutes.GET("/assets/", server.listAssets)
	authRoutes.POST("/asset_categories", server.createAssetCategory)
	authRoutes.GET("/asset_categories/:id", server.getAssetCategory)
	authRoutes.GET("/asset_categories", server.listAssetCategories)
	authRoutes.PUT("/asset_categories", server.updateAssetCategory)
	authRoutes.DELETE("/asset_categories/:id", server.deleteAssetCategory)
	authRoutes.POST("/clients", server.createClient)
	authRoutes.PUT("/clients", server.updateClient)
	authRoutes.GET("/clients/:id", server.getClient)
	authRoutes.GET("/clients/:id/loyalty_total", server.getClientLoyaltyTotal)
	authRoutes.GET("/clients/", server.listClients)
	authRoutes.GET("/clients/by_spending", server.listClientsBySpending)
	authRoutes.DELETE("/clients/:id", server.deleteClient)
	authRoutes.POST("/currencies", server.createCurrency)
	authRoutes.GET("/currencies/:code", server.getCurrency)
	authRoutes.GET("/currencies/", server.listCurrencies)
	authRoutes.PUT("/currencies", server.updateCurrency)
	authRoutes.DELETE("/currencies/:code", server.deleteCurrency)
	authRoutes.POST("/currencies/bulk_delete", server.bulkDeleteCurrencies)

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

	authRoutes.GET("/app_settings", server.getAppSettings)
	authRoutes.PUT("/app_settings", server.updateAppSettings)
	authRoutes.GET("/logs", server.listAppLogs)

	authRoutes.POST("/suppliers", server.createSupplier)
	authRoutes.GET("/suppliers/:id", server.getSupplier)
	authRoutes.GET("/suppliers", server.listSuppliers)

	authRoutes.POST("/colors", server.createColor)
	authRoutes.GET("/colors", server.listColors)
	authRoutes.PUT("/colors", server.updateColor)
	authRoutes.DELETE("/colors/:id", server.deleteColor)
	authRoutes.POST("/colors/bulk_delete", server.bulkDeleteColors)
	authRoutes.POST("/sizes", server.createSize)
	authRoutes.GET("/sizes", server.listSizes)
	authRoutes.PUT("/sizes", server.updateSize)
	authRoutes.DELETE("/sizes/:id", server.deleteSize)
	authRoutes.POST("/sizes/bulk_delete", server.bulkDeleteSizes)

	authRoutes.POST("/sales_invoices", server.createSalesInvoice)
	authRoutes.GET("/sales_invoices/:id", server.getSalesInvoice)
	authRoutes.GET("/sales_invoices", server.listSalesInvoices)

	authRoutes.POST("/return_invoices", server.createReturnInvoice)
	authRoutes.GET("/return_invoices/:id", server.getReturnInvoice)
	authRoutes.GET("/return_invoices", server.listReturnInvoices)

	authRoutes.GET("/invoices/:id/details", server.getInvoiceDetails)
	authRoutes.GET("/branch_invoices/:id/details", server.getBranchInvoiceDetails)

	authRoutes.POST("/invoice_types", server.createInvoiceType)
	authRoutes.GET("/invoice_types/:id", server.getInvoiceType)
	authRoutes.GET("/invoice_types", server.listInvoiceTypes)
	authRoutes.PUT("/invoice_types", server.updateInvoiceType)

	authRoutes.POST("/price_lists", server.createPriceList)
	authRoutes.GET("/price_lists/:id", server.getPriceList)
	authRoutes.GET("/price_lists", server.listPriceLists)
	authRoutes.PUT("/price_lists", server.updatePriceList)
	authRoutes.DELETE("/price_lists/:id", server.deletePriceList)
	authRoutes.POST("/price_lists/items", server.createPriceListItem)
	authRoutes.GET("/price_lists/:id/items", server.listPriceListItems)
	authRoutes.DELETE("/price_lists/:id/items/:product_id", server.deletePriceListItem)
	authRoutes.POST("/price_lists/:id/items/bulk_delete", server.bulkDeletePriceListItems)
	authRoutes.GET("/price_lists/:id/branches", server.listPriceListBranches)
	authRoutes.PUT("/price_lists/:id/branches", server.setPriceListBranches)

	authRoutes.POST("/discount_lists", server.createDiscountList)
	authRoutes.GET("/discount_lists/:id", server.getDiscountList)
	authRoutes.GET("/discount_lists", server.listDiscountLists)
	authRoutes.PUT("/discount_lists", server.updateDiscountList)
	authRoutes.DELETE("/discount_lists/:id", server.deleteDiscountList)
	authRoutes.POST("/discount_lists/items", server.createDiscountListItem)
	authRoutes.GET("/discount_lists/:id/items", server.listDiscountListItems)
	authRoutes.DELETE("/discount_lists/:id/items/:product_id", server.deleteDiscountListItem)
	authRoutes.POST("/discount_lists/:id/items/bulk_delete", server.bulkDeleteDiscountListItems)
	authRoutes.GET("/discount_lists/:id/branches", server.listDiscountListBranches)
	authRoutes.PUT("/discount_lists/:id/branches", server.setDiscountListBranches)

	router.POST("/users/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)
	authRoutes.POST("/users", server.createUser)
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
	authRoutes.PUT("/expenses", server.updateExpense)
	authRoutes.DELETE("/expenses/:id", server.deleteExpense)
	authRoutes.POST("/expenses/bulk_delete", server.bulkDeleteExpenses)
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
	authRoutes.PUT("/loans", server.updateLoan)
	authRoutes.DELETE("/loans/:id", server.deleteLoan)
	authRoutes.POST("/loans/bulk_delete", server.bulkDeleteLoans)

	authRoutes.POST("/loan_payments", server.createLoanPayment)
	authRoutes.GET("/loan_payments", server.listLoanPayments)
	authRoutes.DELETE("/loan_payments/:id", server.deleteLoanPayment)

	authRoutes.POST("/coupon_categories", server.createCouponCategory)
	authRoutes.GET("/coupon_categories/:id", server.getCouponCategory)
	authRoutes.GET("/coupon_categories", server.listCouponCategories)
	authRoutes.PUT("/coupon_categories", server.updateCouponCategory)
	authRoutes.DELETE("/coupon_categories/:id", server.deleteCouponCategory)

	authRoutes.POST("/coupons", server.createCoupon)
	authRoutes.POST("/coupons/bulk_delete", server.bulkDeleteCoupons)
	authRoutes.GET("/coupons/:code", server.getCoupon)
	authRoutes.GET("/coupons", server.listCoupons)
	authRoutes.PUT("/coupons/:code", server.updateCoupon)
	authRoutes.PUT("/coupons/:code/deactivate", server.deactivateCoupon)

	authRoutes.POST("/transfers", server.createTransfer)
	authRoutes.GET("/transfers/:id", server.getTransfer)
	authRoutes.GET("/transfers", server.listTransfers)
	authRoutes.PUT("/transfers", server.updateTransfer)
	authRoutes.POST("/transfers/:id/dispatch", server.dispatchTransfer)
	authRoutes.POST("/transfers/:id/receive", server.receiveTransfer)
	authRoutes.POST("/transfer_items", server.createTransferItem)
	authRoutes.GET("/transfer_items/:transfer_id", server.listTransferItems)

	authRoutes.POST("/purchases", server.createPurchase)
	authRoutes.GET("/purchases/:id", server.getPurchase)
	authRoutes.GET("/purchases", server.listPurchases)
	authRoutes.POST("/purchase_items", server.addPurchaseItem)
	authRoutes.POST("/purchases/:id/receive", server.receivePurchase)

	authRoutes.POST("/branches", server.createBranch)
	authRoutes.GET("/branches", server.listBranches)
	authRoutes.GET("/branches/connected", server.listConnectedBranches)
	authRoutes.POST("/branches/:id/activate", server.setBranchActive(true))
	authRoutes.POST("/branches/:id/deactivate", server.setBranchActive(false))
	authRoutes.POST("/branches/:id/rotate_key", server.rotateBranchKey)

	authRoutes.GET("/branch_expenses", server.listBranchExpenses)
	authRoutes.POST("/branch_expenses", server.adminCreateBranchExpense)
	authRoutes.GET("/branch_expense_images", server.listBranchExpenseImages)
	authRoutes.GET("/branch_expenses/:id/images", server.listBranchExpenseImagesForExpense)
	authRoutes.GET("/branch_expense_images/:id/file", server.getBranchExpenseImageFile)
	authRoutes.GET("/branch_invoices", server.listBranchInvoices)
	authRoutes.GET("/branch_invoices/:id/items", server.listBranchInvoiceItems)
	authRoutes.GET("/branch_invoices/daily_income", server.dailyIncome)
	authRoutes.GET("/branch_shifts", server.listBranchShifts)
	authRoutes.GET("/branch_invoice_settlements", server.listBranchInvoiceSettlements)
	authRoutes.GET("/branch_attendance_changes", server.listBranchAttendanceChanges)
	authRoutes.GET("/branch_attendance_events", server.listBranchAttendanceEvents)
	authRoutes.POST("/branch_attendance_events/import", server.importBranchAttendanceEventsCSV)
	authRoutes.PUT("/branch_attendance_events/:id", server.updateBranchAttendanceEvent)
	authRoutes.DELETE("/branch_attendance_events/:id", server.deleteBranchAttendanceEvent)

	authRoutes.GET("/dashboard/summary", server.getDashboardSummary)
	authRoutes.GET("/dashboard/sales", server.listDashboardSales)
	authRoutes.GET("/dashboard/purchases", server.listDashboardPurchases)
	authRoutes.GET("/dashboard/expenses", server.listDashboardExpenses)
	authRoutes.GET("/dashboard/activities", server.listDashboardActivities)
	authRoutes.GET("/branches/:id/settings", server.getBranchSettings)
	authRoutes.PUT("/branches/:id/settings/managed_locally", server.putBranchSettingsManagedLocally)
	authRoutes.POST("/branches/:id/commands", server.createBranchCommand)
	authRoutes.GET("/branches/:id/commands", server.listBranchCommands)

	authRoutes.GET("/branches/:id/targets", server.listBranchTargets)
	authRoutes.POST("/branches/:id/targets", server.createBranchTarget)
	authRoutes.PUT("/branch_targets/:id", server.updateBranchTarget)
	authRoutes.DELETE("/branch_targets/:id", server.deleteBranchTarget)

	authRoutes.GET("/branches/:id/target_series", server.listBranchTargetSeries)
	authRoutes.POST("/branches/:id/target_series", server.createBranchTargetSeries)
	authRoutes.PUT("/target_series/:id", server.updateBranchTargetSeries)
	authRoutes.PUT("/target_series/:id/active", server.setBranchTargetSeriesActive)
	authRoutes.DELETE("/target_series/:id", server.deleteBranchTargetSeries)

	authRoutes.GET("/branches/:id/salespersons", server.listSalespersons)
	authRoutes.POST("/branches/:id/salespersons", server.createSalesperson)
	authRoutes.PUT("/salespersons/:id", server.updateSalesperson)
	authRoutes.PUT("/salespersons/:id/active", server.setSalespersonActive)

	authRoutes.GET("/employees", server.listEmployees)
	authRoutes.POST("/employees", server.createEmployee)
	authRoutes.PUT("/employees/salaries", server.batchUpdateEmployeeSalaries)
	authRoutes.GET("/employees/:id", server.getEmployee)
	authRoutes.PUT("/employees/:id", server.updateEmployee)
	authRoutes.DELETE("/employees/:id", server.deleteEmployee)

	authRoutes.GET("/branches/:id/branch_users", server.listBranchUsers)
	authRoutes.POST("/branches/:id/branch_users", server.createBranchUser)
	authRoutes.PUT("/branch_users/:id", server.updateBranchUser)
	authRoutes.PUT("/branch_users/:id/pin", server.setBranchUserPin)
	authRoutes.DELETE("/branch_users/:id", server.deleteBranchUser)

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
	branchRoutes.POST("/exchange_invoices", server.createBranchExchangeInvoice)
	branchRoutes.POST("/expenses", server.createBranchExpense)
	branchRoutes.POST("/loans", server.createBranchLoan)
	branchRoutes.POST("/shift_closes", server.createBranchShiftClose)
	branchRoutes.POST("/invoice_settlements", server.createBranchInvoiceSettlement)
	branchRoutes.POST("/attendance_events", server.createBranchAttendanceEvents)
	branchRoutes.POST("/attendance_changes", server.createBranchAttendanceChange)
	branchRoutes.GET("/sync/changes", server.branchSyncChanges)
	branchRoutes.GET("/ws", server.branchWS)
	branchRoutes.PUT("/settings", server.putBranchSettings)
	branchRoutes.POST("/clients", server.putBranchClient)
	branchRoutes.GET("/commands/pending", server.branchPendingCommands)
	branchRoutes.POST("/commands/:id/ack", server.ackBranchCommand)
	branchRoutes.GET("/expenses/upload_status", server.branchExpenseUploadStatus)

	// expense_uploads is deliberately public: a phone scanning the QR code
	// kashi-pos shows has no kashi login and shouldn't need one. Every
	// handler here re-validates its own one-time token instead of relying
	// on any auth middleware -- see expense_upload.go.
	router.GET("/expense_uploads/:token", server.expenseUploadPage)
	router.GET("/expense_uploads/:token/status", server.expenseUploadStatus)
	router.POST("/expense_uploads/:token/images", server.uploadExpenseImages)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
