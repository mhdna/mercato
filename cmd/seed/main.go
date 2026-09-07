package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/lib/pq"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/mhdna/kashi/util"
)

func nullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: true}
}

func nullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: true}
}

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer conn.Close()

	store := db.NewStore(conn)
	ctx := context.Background()

	fmt.Println("🌱 Starting database seeding...")

	// Seed colors
	fmt.Println("📝 Seeding colors...")
	seedColors(ctx, store)

	// Seed suppliers
	fmt.Println("📦 Seeding suppliers...")
	supplierIDs := seedSuppliers(ctx, store)

	// Seed asset types and assets
	fmt.Println("🔧 Seeding asset types and assets...")
	seedAssets(ctx, store)

	// Seed products
	fmt.Println("🛍️ Seeding products...")
	productIDs := seedProducts(ctx, store)

	// Seed one sellable variant (SKU) per product
	fmt.Println("🏷️ Seeding product variants...")
	variantIDs := seedProductVariants(ctx, store, productIDs)

	// Seed currencies
	fmt.Println("💱 Seeding currencies...")
	currencyCodes := seedCurrencies(ctx, store)

	// Seed clients
	fmt.Println("🧑‍🤝‍🧑 Seeding clients...")
	clientIDs := seedClients(ctx, store)

	// Seed cashboxes and cashbox accounts
	fmt.Println("💰 Seeding cashboxes...")
	cashboxIDs := seedCashboxes(ctx, store)
	fmt.Println("🏦 Seeding cashbox accounts...")
	cashboxAccountIDs := seedCashboxAccounts(ctx, store)

	// Seed inventories and stock them with products
	fmt.Println("🏬 Seeding inventories...")
	inventoryIDs := seedInventories(ctx, store)
	fmt.Println("📊 Seeding inventory stock...")
	inventoryStock := seedInventoryStock(ctx, store, inventoryIDs, variantIDs)

	// Seed shifts
	fmt.Println("🕒 Seeding shifts...")
	shifts := seedShifts(ctx, store, cashboxIDs)

	// Seed coupons
	fmt.Println("🎟️ Seeding coupons...")
	seedCoupons(ctx, store, clientIDs)

	// Seed discount lists
	fmt.Println("🏷️ Seeding discount lists...")
	seedDiscountLists(ctx, store, productIDs)

	// Seed transfers
	fmt.Println("🚚 Seeding transfers...")
	seedTransfers(ctx, store, inventoryIDs, variantIDs)

	// Seed purchases
	fmt.Println("🧾 Seeding purchases...")
	seedPurchases(ctx, store, supplierIDs, variantIDs, currencyCodes)

	// Seed sales invoices
	fmt.Println("🧮 Seeding sales invoices...")
	seedSalesInvoices(ctx, store, cashboxIDs, cashboxAccountIDs, shifts, clientIDs, inventoryStock)

	// Seed branch invoices -- the append-only record of what branches sync
	// up, and what the home page's "Recent activity" feed reads from.
	fmt.Println("🏪 Seeding branch invoices...")
	seedBranchInvoices(ctx, conn)

	fmt.Println("✅ Database seeding completed!")
}

var (
	baseColorNames = []string{
		"Red", "Blue", "Green", "Yellow", "Black", "White", "Gray",
		"Purple", "Orange", "Pink", "Brown", "Navy", "Cyan", "Magenta",
	}
	baseColorHexes = []string{
		"#FF0000", "#0000FF", "#00FF00", "#FFFF00", "#000000", "#FFFFFF", "#808080",
		"#800080", "#FFA500", "#FFC0CB", "#8B4513", "#000080", "#00FFFF", "#FF00FF",
	}
	baseSupplierNames = []string{
		"Global Textiles Co.", "Premium Fabrics Ltd.", "Quality Imports Inc.",
		"Fashion Wholesale Hub", "International Traders", "Direct Suppliers LLC",
		"Bulk Goods Distribution", "Elite Manufacturers", "Trade Masters",
		"Universal Supply Chain", "Standard Imports", "Paramount Wholesalers",
	}
	baseProductNames = []string{
		"Cotton T-Shirt", "Denim Jeans", "Casual Shirt", "Summer Dress",
		"Winter Jacket", "Casual Pants", "Polo Shirt", "Tank Top",
		"Hoodie", "Cardigan", "Shorts", "Skirt",
	}
	assetTypes = []string{
		"Furniture", "Equipment", "Electronics", "Vehicles",
		"Tools", "Machinery", "Appliances", "Computer",
	}
)

func seedColors(ctx context.Context, store db.Store) {
	colorVariants := []string{"Light", "Dark", "Bright", "Pale", "Deep", "Soft"}
	count := 0

	for variant, variantPrefix := range colorVariants {
		for i, name := range baseColorNames {
			count++
			colorName := fmt.Sprintf("%s %s", variantPrefix, name)
			// Generate slightly different hex values for variants
			baseHex := baseColorHexes[i]
			adjustedHex := adjustHexColor(baseHex, variant)

			_, err := store.CreateColor(ctx, db.CreateColorParams{
				Name:     colorName,
				HexValue: adjustedHex,
			})
			if err != nil {
				log.Printf("warning: failed to create color %s: %v", colorName, err)
			}
			if count%500 == 0 {
				fmt.Printf("  ✓ Created %d colors\n", count)
			}
		}
	}

	// Add more variations to reach 5000
	for i := 0; i < 5000-(count); i++ {
		colorName := fmt.Sprintf("Custom Color %d", i+1)
		adjustedHex := fmt.Sprintf("#%06X", rand.Intn(16777215))

		_, err := store.CreateColor(ctx, db.CreateColorParams{
			Name:     colorName,
			HexValue: adjustedHex,
		})
		if err != nil {
			log.Printf("warning: failed to create color %s: %v", colorName, err)
		}
		if (i+count+1)%500 == 0 {
			fmt.Printf("  ✓ Created %d colors\n", i+count+1)
		}
	}
	fmt.Printf("  ✓ Completed: 5000 colors created\n")
}

func adjustHexColor(hex string, variant int) string {
	// Generate consistent variants of hex colors
	baseVal := rand.Intn(50) + (variant * 20)
	return fmt.Sprintf("#%06X", baseVal*100000)
}

func seedSuppliers(ctx context.Context, store db.Store) []int64 {
	countries := []string{"USA", "China", "India", "Vietnam", "Bangladesh", "Turkey", "Mexico", "Brazil", "Germany", "Japan"}
	supplierIDs := make([]int64, 0, 5000)

	for i := 0; i < 5000; i++ {
		baseName := baseSupplierNames[i%len(baseSupplierNames)]
		name := fmt.Sprintf("%s - %d", baseName, i+1)
		country := countries[rand.Intn(len(countries))]
		phone := fmt.Sprintf("+%d%010d", rand.Intn(999)+1, rand.Intn(10000000000))
		lat := rand.Float64()*180 - 90
		lng := rand.Float64()*360 - 180

		supplier, err := store.CreateSupplier(ctx, db.CreateSupplierParams{
			Name:             name,
			Phone:            phone,
			Country:          country,
			Address:          fmt.Sprintf("%d Main St, City", rand.Intn(9999)+1),
			AddressLatitude:  nullFloat64(lat),
			AddressLongitude: nullFloat64(lng),
		})
		if err != nil {
			log.Printf("warning: failed to create supplier %s: %v", name, err)
		} else {
			supplierIDs = append(supplierIDs, supplier.ID)
		}
		if (i+1)%500 == 0 {
			fmt.Printf("  ✓ Created %d suppliers\n", i+1)
		}
	}
	fmt.Printf("  ✓ Completed: %d/5000 suppliers created\n", len(supplierIDs))

	// If most/all inserts collided with an earlier run of this seeder (unique
	// name/code constraints), fall back to whatever suppliers already exist
	// so entities that reference a supplier ID still have something to use.
	if len(supplierIDs) < 500 {
		existing := fetchExistingSupplierIDs(ctx, store, 1000)
		fmt.Printf("  ↺ Reusing %d pre-existing suppliers\n", len(existing))
		supplierIDs = append(supplierIDs, existing...)
	}
	return supplierIDs
}

func fetchExistingSupplierIDs(ctx context.Context, store db.Store, limit int) []int64 {
	ids := make([]int64, 0, limit)
	const pageSize = 100
	for offset := int32(0); len(ids) < limit; offset += pageSize {
		suppliers, err := store.ListSuppliers(ctx, db.ListSuppliersParams{PageSize: pageSize, PageOffset: offset})
		if err != nil || len(suppliers) == 0 {
			break
		}
		for _, s := range suppliers {
			ids = append(ids, s.ID)
			if len(ids) >= limit {
				break
			}
		}
	}
	return ids
}

func seedAssets(ctx context.Context, store db.Store) {
	// Create asset categories
	var assetCategoryIDs []int64
	for _, aType := range assetTypes {
		category, err := store.CreateAssetCategory(ctx, db.CreateAssetCategoryParams{
			Name:     aType,
			IsActive: true,
			Icon:     "mdi-tag-outline",
			Color:    "blue-grey",
		})
		if err != nil {
			log.Printf("warning: failed to create asset category %s: %v", aType, err)
		} else {
			assetCategoryIDs = append(assetCategoryIDs, category.ID)
			fmt.Printf("  ✓ Created asset category: %s\n", aType)
		}
	}

	// Create 5000 assets
	for i := 0; i < 5000; i++ {
		if len(assetCategoryIDs) == 0 {
			break
		}
		categoryID := assetCategoryIDs[rand.Intn(len(assetCategoryIDs))]
		name := fmt.Sprintf("Asset %d", i+1)
		code := fmt.Sprintf("AST%05d", 10000+i)
		boughtAt := time.Now().AddDate(-rand.Intn(5), -rand.Intn(12), -rand.Intn(30))

		_, err := store.CreateAsset(ctx, db.CreateAssetParams{
			Name:       name,
			Code:       code,
			CategoryID: categoryID,
			BoughtAt:   boughtAt,
		})
		if err != nil {
			log.Printf("warning: failed to create asset %s: %v", name, err)
		}
		if (i+1)%500 == 0 {
			fmt.Printf("  ✓ Created %d assets\n", i+1)
		}
	}
	fmt.Printf("  ✓ Completed: 5000 assets created\n")
}

func seedProducts(ctx context.Context, store db.Store) []int64 {
	productIDs := make([]int64, 0, 5000)

	for i := 0; i < 5000; i++ {
		baseProduct := baseProductNames[i%len(baseProductNames)]
		name := fmt.Sprintf("%s - Variant %d", baseProduct, i/len(baseProductNames)+1)
		code := fmt.Sprintf("PRD%05d", 10000+i)
		description := fmt.Sprintf("High-quality %s with excellent craftsmanship. Variant: %d", baseProduct, i/len(baseProductNames)+1)

		product, err := store.CreateProduct(ctx, db.CreateProductParams{
			Code:        code,
			Name:        name,
			Description: description,
		})
		if err != nil {
			log.Printf("warning: failed to create product %s: %v", name, err)
		} else {
			productIDs = append(productIDs, product.ID)
		}
		if (i+1)%500 == 0 {
			fmt.Printf("  ✓ Created %d products\n", i+1)
		}
	}
	fmt.Printf("  ✓ Completed: %d/5000 products created\n", len(productIDs))

	// Same fallback as suppliers: reuse existing products if this seeder has
	// already run before and every code collided.
	if len(productIDs) < 500 {
		existing := fetchExistingProductIDs(ctx, store, 1000)
		fmt.Printf("  ↺ Reusing %d pre-existing products\n", len(existing))
		productIDs = append(productIDs, existing...)
	}
	return productIDs
}

func fetchExistingProductIDs(ctx context.Context, store db.Store, limit int) []int64 {
	ids := make([]int64, 0, limit)
	const pageSize = 100
	for offset := int32(0); len(ids) < limit; offset += pageSize {
		products, err := store.ListProducts(ctx, db.ListProductsParams{PageLimit: pageSize, PageOffset: offset})
		if err != nil || len(products) == 0 {
			break
		}
		for _, p := range products {
			ids = append(ids, p.ID)
			if len(ids) >= limit {
				break
			}
		}
	}
	return ids
}

func seedCurrencies(ctx context.Context, store db.Store) []string {
	// USC and USD already come from the migration seed data; add a few more.
	extra := []struct {
		name   string
		code   string
		symbol string
		value  int64
	}{
		{"Euro Cent", "EUC", "€", 92},
		{"British Pence", "GBP", "£", 79},
		{"Lebanese Pound", "LBP", "ل.ل", 8950000},
		{"Canadian Cent", "CAC", "C$", 136},
	}

	codes := []string{"USC", "USD"}
	for _, c := range extra {
		_, err := store.CreateCurrency(ctx, db.CreateCurrencyParams{
			Name:                   c.name,
			Code:                   c.code,
			Symbol:                 c.symbol,
			ValueInDefaultCurrency: c.value,
		})
		if err != nil {
			log.Printf("warning: failed to create currency %s: %v", c.code, err)
			continue
		}
		codes = append(codes, c.code)
	}
	fmt.Printf("  ✓ Completed: %d currencies available\n", len(codes))
	return codes
}

var (
	firstNames = []string{"Aisha", "Omar", "Layla", "Karim", "Nour", "Ziad", "Maya", "Youssef", "Rana", "Hassan"}
	lastNames  = []string{"Khalil", "Haddad", "Saleh", "Mansour", "Nasser", "Fares", "Aziz", "Farah", "Sabbagh", "Rahal"}
)

func seedClients(ctx context.Context, store db.Store) []int64 {
	clientIDs := make([]int64, 0, 200)
	for i := 0; i < 200; i++ {
		name := fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))])
		phone := fmt.Sprintf("%d%07d", rand.Intn(9)+1, rand.Intn(10000000))

		client, err := store.CreateClient(ctx, db.CreateClientParams{
			Name:  name,
			Phone: phone,
		})
		if err != nil {
			log.Printf("warning: failed to create client %s: %v", name, err)
			continue
		}
		clientIDs = append(clientIDs, client.ID)
	}
	fmt.Printf("  ✓ Completed: %d clients created\n", len(clientIDs))
	return clientIDs
}

func seedCashboxes(ctx context.Context, store db.Store) []int64 {
	names := []string{"Main Register", "Front Desk", "Warehouse Office", "Downtown Branch", "Online Orders"}
	cashboxIDs := make([]int64, 0, len(names))
	for i, name := range names {
		cashbox, err := store.CreateCashbox(ctx, db.CreateCashboxParams{
			Code:     fmt.Sprintf("CB%03d", i+1),
			Name:     name,
			IsActive: true,
		})
		if err != nil {
			log.Printf("warning: failed to create cashbox %s: %v", name, err)
			continue
		}
		cashboxIDs = append(cashboxIDs, cashbox.ID)
	}
	fmt.Printf("  ✓ Completed: %d cashboxes created\n", len(cashboxIDs))
	if len(cashboxIDs) == 0 {
		existing := fetchExistingCashboxIDs(ctx, store, len(names))
		fmt.Printf("  ↺ Reusing %d pre-existing cashboxes\n", len(existing))
		cashboxIDs = append(cashboxIDs, existing...)
	}
	return cashboxIDs
}

func fetchExistingCashboxIDs(ctx context.Context, store db.Store, limit int) []int64 {
	cashboxes, err := store.ListCashboxes(ctx, db.ListCashboxesParams{Limit: int32(limit), Offset: 0})
	if err != nil {
		return nil
	}
	ids := make([]int64, 0, len(cashboxes))
	for _, c := range cashboxes {
		ids = append(ids, c.ID)
	}
	return ids
}

func seedCashboxAccounts(ctx context.Context, store db.Store) []int64 {
	names := []string{"Cash", "Card", "Bank Transfer", "Mobile Wallet"}
	accountIDs := make([]int64, 0, len(names))
	for _, name := range names {
		account, err := store.CreateCashboxAccount(ctx, db.CreateCashboxAccountParams{Name: name, CurrencyCode: "USD"})
		if err != nil {
			log.Printf("warning: failed to create cashbox account %s: %v", name, err)
			continue
		}
		accountIDs = append(accountIDs, account.ID)
	}
	fmt.Printf("  ✓ Completed: %d cashbox accounts created\n", len(accountIDs))
	if len(accountIDs) == 0 {
		existing := fetchExistingCashboxAccountIDs(ctx, store, len(names))
		fmt.Printf("  ↺ Reusing %d pre-existing cashbox accounts\n", len(existing))
		accountIDs = append(accountIDs, existing...)
	}
	return accountIDs
}

func fetchExistingCashboxAccountIDs(ctx context.Context, store db.Store, limit int) []int64 {
	accounts, err := store.ListCashboxAccounts(ctx, db.ListCashboxAccountsParams{Limit: int32(limit), Offset: 0})
	if err != nil {
		return nil
	}
	ids := make([]int64, 0, len(accounts))
	for _, a := range accounts {
		ids = append(ids, a.ID)
	}
	return ids
}

func seedInventories(ctx context.Context, store db.Store) []int64 {
	inventories := []struct {
		name string
		typ  db.InventoryType
		code string
	}{
		{"Main Warehouse", db.InventoryTypeWarehouse, "WH-MAIN"},
		{"Overflow Warehouse", db.InventoryTypeWarehouse, "WH-OVERFLOW"},
		{"Downtown Store", db.InventoryTypeStore, "ST-DOWNTOWN"},
		{"Mall Store", db.InventoryTypeStore, "ST-MALL"},
		{"Airport Store", db.InventoryTypeStore, "ST-AIRPORT"},
	}

	inventoryIDs := make([]int64, 0, len(inventories))
	for _, inv := range inventories {
		lat := rand.Float64()*180 - 90
		lng := rand.Float64()*360 - 180
		inventory, err := store.CreateInventory(ctx, db.CreateInventoryParams{
			Name:      inv.name,
			Type:      inv.typ,
			Code:      inv.code,
			Latitude:  nullFloat64(lat),
			Longitude: nullFloat64(lng),
		})
		if err != nil {
			log.Printf("warning: failed to create inventory %s: %v", inv.name, err)
			continue
		}
		inventoryIDs = append(inventoryIDs, inventory.ID)
	}
	fmt.Printf("  ✓ Completed: %d inventories created\n", len(inventoryIDs))
	if len(inventoryIDs) == 0 {
		existing := fetchExistingInventoryIDs(ctx, store, len(inventories))
		fmt.Printf("  ↺ Reusing %d pre-existing inventories\n", len(existing))
		inventoryIDs = append(inventoryIDs, existing...)
	}
	return inventoryIDs
}

func fetchExistingInventoryIDs(ctx context.Context, store db.Store, limit int) []int64 {
	inventories, err := store.ListInventories(ctx, db.ListInventoriesParams{Limit: int32(limit), Offset: 0})
	if err != nil {
		return nil
	}
	ids := make([]int64, 0, len(inventories))
	for _, inv := range inventories {
		ids = append(ids, inv.ID)
	}
	return ids
}

// seedProductVariants creates one sellable variant (SKU) per product, so
// stock, transfers, purchases and sales invoices -- all variant-keyed now
// -- have something to reference.
func seedProductVariants(ctx context.Context, store db.Store, productIDs []int64) []int64 {
	variantIDs := make([]int64, 0, len(productIDs))
	for _, productID := range productIDs {
		variant, err := store.CreateProductVariant(ctx, db.CreateProductVariantParams{
			ProductID: productID,
			Barcode:   fmt.Sprintf("VAR%010d", productID),
			Price:     nullInt64(int64(rand.Intn(9000) + 1000)),
		})
		if err != nil {
			log.Printf("warning: failed to create variant for product %d: %v", productID, err)
			continue
		}
		variantIDs = append(variantIDs, variant.ID)
	}
	fmt.Printf("  ✓ Completed: %d variants created\n", len(variantIDs))
	return variantIDs
}

// seedInventoryStock stocks a random sample of variants in each inventory and
// returns, per inventory, the variant IDs that now have stock there - useful
// for seeding sales invoices/transfers without going negative on quantity.
func seedInventoryStock(ctx context.Context, store db.Store, inventoryIDs, variantIDs []int64) map[int64][]int64 {
	stock := make(map[int64][]int64, len(inventoryIDs))
	sampleSize := 300
	if sampleSize > len(variantIDs) {
		sampleSize = len(variantIDs)
	}

	for _, invID := range inventoryIDs {
		perm := rand.Perm(len(variantIDs))[:sampleSize]
		stocked := make([]int64, 0, sampleSize)
		for _, idx := range perm {
			variantID := variantIDs[idx]
			qty := int64(rand.Intn(900) + 100)
			_, err := store.AddInventoryStockQuantity(ctx, db.AddInventoryStockQuantityParams{
				InventoryID: invID,
				VariantID:   variantID,
				Quantity:    qty,
			})
			if err != nil {
				log.Printf("warning: failed to stock variant %d in inventory %d: %v", variantID, invID, err)
				continue
			}
			stocked = append(stocked, variantID)
		}
		stock[invID] = stocked
	}
	fmt.Printf("  ✓ Completed: stocked %d inventories with up to %d variants each\n", len(inventoryIDs), sampleSize)
	return stock
}

func seedShifts(ctx context.Context, store db.Store, cashboxIDs []int64) []db.Shift {
	shifts := make([]db.Shift, 0, len(cashboxIDs)*3)
	for _, cashboxID := range cashboxIDs {
		for i := 0; i < 3; i++ {
			shift, err := store.CreateShift(ctx, cashboxID)
			if err != nil {
				log.Printf("warning: failed to create shift for cashbox %d: %v", cashboxID, err)
				continue
			}
			// Close about a third of the shifts, leave the rest open for invoices.
			if i == 0 {
				if err := store.CloseShift(ctx, shift.ID); err != nil {
					log.Printf("warning: failed to close shift %d: %v", shift.ID, err)
				} else {
					shift.IsClosed = true
				}
			}
			shifts = append(shifts, shift)
		}
	}
	fmt.Printf("  ✓ Completed: %d shifts created\n", len(shifts))
	return shifts
}

func seedCoupons(ctx context.Context, store db.Store, clientIDs []int64) {
	if len(clientIDs) == 0 {
		return
	}
	discountTypes := []db.DiscountType{db.DiscountTypeFixed, db.DiscountTypePercentage}
	reasons := []string{"Loyalty reward", "Birthday promo", "Referral bonus", "Seasonal sale", "Customer complaint compensation"}

	count := 0
	for i := 0; i < 100; i++ {
		clientID := clientIDs[rand.Intn(len(clientIDs))]
		code := fmt.Sprintf("COUPON%05d", i+1)
		status := db.CouponStatusActive
		if rand.Intn(4) == 0 {
			status = db.CouponStatusInactive
		}

		_, err := store.CreateCoupon(ctx, db.CreateCouponParams{
			Code:         code,
			Status:       status,
			DiscountType: discountTypes[rand.Intn(len(discountTypes))],
			Reason:       reasons[rand.Intn(len(reasons))],
			ClientID:     clientID,
			ValidUntil:   time.Now().AddDate(0, rand.Intn(12)+1, 0),
		})
		if err != nil {
			log.Printf("warning: failed to create coupon %s: %v", code, err)
			continue
		}
		count++
	}
	fmt.Printf("  ✓ Completed: %d coupons created\n", count)
}

func seedDiscountLists(ctx context.Context, store db.Store, productIDs []int64) {
	if len(productIDs) == 0 {
		return
	}
	lists := []string{"Summer Sale", "Clearance", "VIP Discount", "Black Friday", "New Season Launch"}

	listCount, itemCount := 0, 0
	for i, name := range lists {
		list, err := store.CreateDiscountList(ctx, db.CreateDiscountListParams{
			Name:      name,
			IsActive:  true,
			IsDefault: i == 0,
			ValidFrom: time.Now().AddDate(0, -1, 0),
			ValidTo:   time.Now().AddDate(0, 3, 0),
		})
		if err != nil {
			log.Printf("warning: failed to create discount list %s: %v", name, err)
			continue
		}
		listCount++

		perm := rand.Perm(len(productIDs))
		itemsForList := 20
		if itemsForList > len(perm) {
			itemsForList = len(perm)
		}
		for _, idx := range perm[:itemsForList] {
			_, err := store.CreateDiscountListItem(ctx, db.CreateDiscountListItemParams{
				DiscountListID: list.ID,
				ProductID:      productIDs[idx],
				Discount:       int16(rand.Intn(40) + 5),
			})
			if err != nil {
				log.Printf("warning: failed to add product %d to discount list %d: %v", productIDs[idx], list.ID, err)
				continue
			}
			itemCount++
		}
	}
	fmt.Printf("  ✓ Completed: %d discount lists with %d items created\n", listCount, itemCount)
}

func seedTransfers(ctx context.Context, store db.Store, inventoryIDs, variantIDs []int64) {
	if len(inventoryIDs) < 2 || len(variantIDs) == 0 {
		return
	}

	transferCount, itemCount := 0, 0
	for i := 0; i < 40; i++ {
		from := inventoryIDs[rand.Intn(len(inventoryIDs))]
		to := inventoryIDs[rand.Intn(len(inventoryIDs))]
		if from == to {
			continue
		}

		items := make([]db.TransferItemParams, 0, 4)
		for j := 0; j < rand.Intn(4)+1; j++ {
			items = append(items, db.TransferItemParams{
				VariantID: nullInt64(variantIDs[rand.Intn(len(variantIDs))]),
				Quantity:  int64(rand.Intn(20) + 1),
			})
		}

		result, err := store.CreateTransferTx(ctx, db.CreateTransferTxParams{
			FromInventoryID: from,
			ToInventoryID:   to,
			Type:            db.TransferTypeProducts,
			Items:           items,
		})
		if err != nil {
			log.Printf("warning: failed to create transfer: %v", err)
			continue
		}
		transferCount++
		itemCount += len(result.Items)
	}
	fmt.Printf("  ✓ Completed: %d transfers with %d items created\n", transferCount, itemCount)
}

func seedPurchases(ctx context.Context, store db.Store, supplierIDs, variantIDs []int64, currencyCodes []string) {
	if len(supplierIDs) == 0 || len(variantIDs) == 0 || len(currencyCodes) == 0 {
		return
	}

	purchaseCount, itemCount := 0, 0
	for i := 0; i < 60; i++ {
		items := make([]db.PurchaseItemParams, 0, 5)
		for j := 0; j < rand.Intn(5)+1; j++ {
			items = append(items, db.PurchaseItemParams{
				VariantID: nullInt64(variantIDs[rand.Intn(len(variantIDs))]),
				Quantity:  int64(rand.Intn(100) + 10),
				UnitPrice: int64(rand.Intn(5000) + 100),
			})
		}

		result, err := store.CreatePurchaseTx(ctx, db.CreatePurchaseTxParams{
			SupplierID:   supplierIDs[rand.Intn(len(supplierIDs))],
			CurrencyCode: currencyCodes[rand.Intn(len(currencyCodes))],
			PurchasedAt:  sql.NullTime{Time: time.Now().AddDate(0, 0, -rand.Intn(365)), Valid: true},
			Items:        items,
		})
		if err != nil {
			log.Printf("warning: failed to create purchase: %v", err)
			continue
		}
		purchaseCount++
		itemCount += len(result.Items)
	}
	fmt.Printf("  ✓ Completed: %d purchases with %d items created\n", purchaseCount, itemCount)
}

// seedBranchInvoices fills branch_invoices (plus a couple of items and a
// payment row each) with 1000 rows spread across the branches and across
// the current calendar month. received_at is staggered on purpose: the
// dashboard's activity feed orders and range-filters on received_at, so
// evenly-spaced values make the "Recent activity" list look real.
//
// It writes with raw SQL rather than the CreateBranchInvoice query because
// that query has no received_at parameter (it defaults to now()), and here
// we specifically want it backdated.
func seedBranchInvoices(ctx context.Context, conn *sql.DB) {
	branchIDs := ensureSeedBranches(ctx, conn)
	if len(branchIDs) == 0 {
		log.Println("warning: no branches available, skipping branch invoices")
		return
	}

	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	window := now.Sub(monthStart)
	if window < time.Hour {
		window = time.Hour
	}

	salespeople := []string{"Aisha K.", "Omar H.", "Layla S.", "Karim M.", "Nour F.", "Ziad A."}
	// Weighted toward plain sales, with some returns and exchanges mixed in.
	kinds := []string{"sales", "sales", "sales", "sales", "return", "exchange"}
	accounts := []string{"Cash", "Card", "Mobile Wallet"}

	const total = 1000
	created := 0
	for i := 0; i < total; i++ {
		branchID := branchIDs[i%len(branchIDs)]
		kind := kinds[rand.Intn(len(kinds))]

		itemsTotal := int64(rand.Intn(24000) + 1000) // $10.00 - $250.00 in cents
		discount := 0
		if rand.Intn(3) == 0 {
			discount = rand.Intn(20) + 5
		}
		grandTotal := itemsTotal - itemsTotal*int64(discount)/100
		switch kind {
		case "return":
			grandTotal = -grandTotal // a refund reads as negative in the feed
		case "exchange":
			grandTotal = int64(rand.Intn(6000)) - 3000 // small signed net difference
		}

		occurredAt := now.Add(-time.Duration(rand.Int63n(int64(window))))
		receivedAt := occurredAt.Add(time.Duration(rand.Intn(90)) * time.Minute)
		if receivedAt.After(now) {
			receivedAt = now
		}

		clientRef := fmt.Sprintf("SEED-%06d", i+1)
		code := fmt.Sprintf("BINV-%06d", 100000+i)

		var invoiceID int64
		err := conn.QueryRowContext(ctx, `
			INSERT INTO branch_invoices (
				branch_id, client_ref, kind, branch_invoice_code,
				branch_cashbox_account_id, branch_shift_id, branch_inventory_id,
				branch_client_id, related_client_ref,
				discount, subtotal, discounted_total, grand_total,
				loyalty_points_delta, occurred_at, received_at, salesperson_name
			) VALUES (
				$1, $2, $3, $4,
				$5, $6, $7,
				NULL, NULL,
				$8, $9, $10, $11,
				$12, $13, $14, $15
			)
			ON CONFLICT (branch_id, client_ref) DO NOTHING
			RETURNING id`,
			branchID, clientRef, kind, code,
			rand.Int63n(5)+1, rand.Int63n(9)+1, rand.Int63n(5)+1,
			discount, itemsTotal, itemsTotal, grandTotal,
			grandTotal/100, occurredAt, receivedAt,
			salespeople[rand.Intn(len(salespeople))],
		).Scan(&invoiceID)
		if err == sql.ErrNoRows {
			continue // already seeded on an earlier run
		}
		if err != nil {
			log.Printf("warning: failed to insert branch invoice %d: %v", i, err)
			continue
		}

		for j := 0; j < rand.Intn(2)+1; j++ {
			qty := int64(rand.Intn(3) + 1)
			unitPrice := int64(rand.Intn(8000) + 500)
			_, err := conn.ExecContext(ctx, `
				INSERT INTO branch_invoice_items (
					branch_invoice_id, branch_product_id, unit_price, line_total, discount, quantity
				) VALUES ($1, $2, $3, $4, $5, $6)`,
				invoiceID, rand.Int63n(500)+1, unitPrice, unitPrice*qty, 0, qty)
			if err != nil {
				log.Printf("warning: failed to insert branch invoice item: %v", err)
			}
		}

		if _, err := conn.ExecContext(ctx, `
			INSERT INTO branch_invoice_payments (branch_invoice_id, account_name, amount)
			VALUES ($1, $2, $3)`,
			invoiceID, accounts[rand.Intn(len(accounts))], grandTotal,
		); err != nil {
			log.Printf("warning: failed to insert branch invoice payment: %v", err)
		}

		created++
		if created%200 == 0 {
			fmt.Printf("  ✓ Created %d branch invoices\n", created)
		}
	}
	fmt.Printf("  ✓ Completed: %d branch invoices created\n", created)
}

// ensureSeedBranches returns existing branch IDs, creating a small set of
// placeholder branches first if the table is empty.
func ensureSeedBranches(ctx context.Context, conn *sql.DB) []int64 {
	if ids := queryInt64s(ctx, conn, `SELECT id FROM branches ORDER BY id`); len(ids) > 0 {
		return ids
	}

	names := []string{"Beirut Central", "Tripoli Depot", "Saida Store", "Zahle Hub", "Jounieh Outlet"}
	ids := make([]int64, 0, len(names))
	for i, name := range names {
		var id int64
		err := conn.QueryRowContext(ctx, `
			INSERT INTO branches (name, code, api_key_hash)
			VALUES ($1, $2, $3)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
			RETURNING id`,
			name, fmt.Sprintf("BR-SEED-%02d", i+1), "seed-placeholder-not-a-real-key",
		).Scan(&id)
		if err != nil {
			log.Printf("warning: failed to ensure branch %s: %v", name, err)
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func queryInt64s(ctx context.Context, conn *sql.DB, query string) []int64 {
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return out
		}
		out = append(out, v)
	}
	return out
}

// seedSalesInvoices creates invoices whose totals are computed with the exact
// same arithmetic as db.validateInvoiceAmounts, since SalesInvoiceTx rejects
// anything that doesn't add up.
func seedSalesInvoices(
	ctx context.Context,
	store db.Store,
	cashboxIDs, cashboxAccountIDs []int64,
	shifts []db.Shift,
	clientIDs []int64,
	inventoryStock map[int64][]int64,
) {
	if len(cashboxAccountIDs) == 0 || len(clientIDs) == 0 || len(shifts) == 0 {
		return
	}

	// Only use open shifts, grouped by cashbox, so cashbox/shift pairs line up.
	shiftsByCashbox := make(map[int64][]int64)
	for _, s := range shifts {
		if s.IsClosed {
			continue
		}
		shiftsByCashbox[s.CashboxID] = append(shiftsByCashbox[s.CashboxID], s.ID)
	}

	inventoryIDs := make([]int64, 0, len(inventoryStock))
	for invID, products := range inventoryStock {
		if len(products) > 0 {
			inventoryIDs = append(inventoryIDs, invID)
		}
	}
	if len(inventoryIDs) == 0 {
		return
	}

	count := 0
	for i := 0; i < 150; i++ {
		cashboxID := cashboxIDs[rand.Intn(len(cashboxIDs))]
		shiftIDs := shiftsByCashbox[cashboxID]
		if len(shiftIDs) == 0 {
			continue
		}
		shiftID := shiftIDs[rand.Intn(len(shiftIDs))]

		inventoryID := inventoryIDs[rand.Intn(len(inventoryIDs))]
		stockedProducts := inventoryStock[inventoryID]
		if len(stockedProducts) == 0 {
			continue
		}

		itemsForInvoice := rand.Intn(4) + 1
		items := make([]db.SalesInvoiceItem, 0, itemsForInvoice)
		var itemsTotal, netTotal int64
		for j := 0; j < itemsForInvoice; j++ {
			variantID := stockedProducts[rand.Intn(len(stockedProducts))]
			quantity := int64(rand.Intn(3) + 1)
			unitPrice := int64(rand.Intn(4000) + 500)
			discount := int16(0)
			if rand.Intn(3) == 0 {
				discount = int16(rand.Intn(20) + 5)
			}

			itemPrice := unitPrice * quantity
			itemDiscounted := itemPrice
			if discount > 0 {
				itemDiscounted = itemPrice - (itemPrice * int64(discount) / 100)
			}
			itemsTotal += itemPrice
			netTotal += itemDiscounted

			items = append(items, db.SalesInvoiceItem{
				VariantID: variantID,
				UnitPrice: unitPrice,
				LineTotal: itemDiscounted,
				Discount:  discount,
				Quantity:  quantity,
			})
		}

		invoiceDiscount := int16(0)
		if rand.Intn(4) == 0 {
			invoiceDiscount = int16(rand.Intn(10) + 5)
		}
		subTotal := itemsTotal * (100 - int64(invoiceDiscount)) / 100

		_, err := store.SalesInvoiceTx(ctx, db.SalesInvoiceTxParams{
			CashboxID:        cashboxID,
			ShiftID:          shiftID,
			Year:             int32(time.Now().Year()),
			ClientID:         clientIDs[rand.Intn(len(clientIDs))],
			InventoryID:      inventoryID,
			Discount:         invoiceDiscount,
			SubTotal:         subTotal,
			DiscountedTotal:  itemsTotal,
			GrandTotal:       netTotal,
			Items:            items,
			CashboxAccountID: cashboxAccountIDs[rand.Intn(len(cashboxAccountIDs))],
		})
		if err != nil {
			log.Printf("warning: failed to create sales invoice: %v", err)
			continue
		}
		count++
	}
	fmt.Printf("  ✓ Completed: %d sales invoices created\n", count)
}
