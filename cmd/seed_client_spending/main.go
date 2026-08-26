// seed_client_spending fills the database with fake clients (retail and
// wholesale) plus branch-reported sales, so the clients "loyalty pyramid"
// page has a realistic-looking spending distribution to preview locally.
// It's deliberately separate from cmd/seed (which reseeds thousands of
// unrelated rows like colors/products/suppliers) -- run with:
//
//	go run ./cmd/seed_client_spending
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

var (
	firstNames = []string{
		"Aisha", "Omar", "Layla", "Karim", "Nour", "Ziad", "Maya", "Youssef", "Rana", "Hassan",
		"Tarek", "Salma", "Jad", "Dana", "Rami", "Farah", "Bassel", "Lina", "Sami", "Hala",
		"Nadim", "Yasmin", "Marwan", "Reem", "Fadi",
	}
	lastNames = []string{
		"Khalil", "Haddad", "Saleh", "Mansour", "Nasser", "Fares", "Aziz", "Farah", "Sabbagh", "Rahal",
		"Choueiri", "Barakat", "Younes", "Karam", "Nassif", "Chami", "Abou Diab", "Zaher",
	}
	businessPrefixes = []string{
		"Al Noor", "Cedar", "Beirut", "Golden", "Silver Star", "Union", "Metro", "Prime",
		"Levant", "Coastal", "Highland", "Sunrise",
	}
	businessSuffixes = []string{
		"Trading", "Boutique", "Fashion House", "Wholesale Co.", "Textiles", "Retail Group",
		"Market", "Emporium",
	}
)

// tier is a target spend band, weighted so the overall distribution looks
// like a pyramid: a handful of whales, progressively more clients the
// smaller their spend gets.
type tier struct {
	label    string
	minCents int64
	maxCents int64
	weight   int
}

var tiers = []tier{
	{"diamond", 500_000, 1_200_000, 3}, // $5,000 - $12,000
	{"platinum", 200_000, 500_000, 7},  // $2,000 - $5,000
	{"gold", 80_000, 200_000, 15},      // $800 - $2,000
	{"silver", 30_000, 80_000, 30},     // $300 - $800
	{"bronze", 0, 30_000, 45},          // $0 - $300
}

func pickTier() tier {
	total := 0
	for _, t := range tiers {
		total += t.weight
	}
	roll := rand.Intn(total)
	for _, t := range tiers {
		if roll < t.weight {
			return t
		}
		roll -= t.weight
	}
	return tiers[len(tiers)-1]
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

	fmt.Println("🌱 Seeding clients + branch spending...")

	branchIDs := ensureBranches(ctx, store)
	fmt.Printf("  ✓ Using %d branches\n", len(branchIDs))

	retailCount := seedClientsWithSpending(ctx, store, "retail", 55, branchIDs)
	wholesaleCount := seedClientsWithSpending(ctx, store, "wholesale", 20, branchIDs)

	fmt.Printf("✅ Done: %d retail + %d wholesale clients seeded with spending history\n", retailCount, wholesaleCount)
}

func ensureBranches(ctx context.Context, store db.Store) []int64 {
	existing, err := store.ListBranches(ctx)
	if err != nil {
		log.Fatal("cannot list branches:", err)
	}

	ids := make([]int64, 0, len(existing)+2)
	for _, b := range existing {
		ids = append(ids, b.ID)
	}

	names := []string{"Downtown Branch", "Mall Branch", "Airport Branch"}
	for _, name := range names {
		if len(ids) >= 3 {
			break
		}
		code := fmt.Sprintf("SEED-%s", name)
		branch, err := store.CreateBranch(ctx, db.CreateBranchParams{
			Name:       name,
			Code:       code,
			ApiKeyHash: "seed-placeholder-hash",
		})
		if err != nil {
			log.Printf("warning: failed to create branch %s (may already exist): %v", name, err)
			continue
		}
		ids = append(ids, branch.ID)
	}
	return ids
}

func randomName(clientType string) string {
	if clientType == "wholesale" {
		return fmt.Sprintf("%s %s", businessPrefixes[rand.Intn(len(businessPrefixes))], businessSuffixes[rand.Intn(len(businessSuffixes))])
	}
	return fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))])
}

func seedClientsWithSpending(ctx context.Context, store db.Store, clientType string, count int, branchIDs []int64) int {
	if len(branchIDs) == 0 {
		log.Println("warning: no branches available, skipping spending seed")
		return 0
	}

	created := 0
	for i := 0; i < count; i++ {
		name := randomName(clientType)
		phone := fmt.Sprintf("%d%08d", rand.Intn(9)+1, rand.Intn(100000000))

		client, err := store.CreateClient(ctx, db.CreateClientParams{
			Name:       name,
			Phone:      phone,
			ClientType: clientType,
		})
		if err != nil {
			log.Printf("warning: failed to create client %s: %v", name, err)
			continue
		}
		created++

		target := pickTier()
		targetSpend := target.minCents + rand.Int63n(target.maxCents-target.minCents+1)

		// Spend at one or two branches, split unevenly so the branch filter
		// actually shows a different ranking than "All Branches".
		usedBranches := []int64{branchIDs[rand.Intn(len(branchIDs))]}
		if len(branchIDs) > 1 && rand.Intn(3) == 0 {
			second := branchIDs[rand.Intn(len(branchIDs))]
			if second != usedBranches[0] {
				usedBranches = append(usedBranches, second)
			}
		}

		remaining := targetSpend
		for bi, branchID := range usedBranches {
			branchShare := remaining
			if bi == 0 && len(usedBranches) > 1 {
				branchShare = int64(float64(remaining) * (0.6 + rand.Float64()*0.25))
			}
			seedBranchSpend(ctx, store, branchID, client.ID, branchShare)
			remaining -= branchShare
		}

		if created%10 == 0 {
			fmt.Printf("  ✓ Seeded %d %s clients\n", created, clientType)
		}
	}
	return created
}

// seedBranchSpend records `amountCents` of net spend for one client at one
// branch, spread across a handful of sales invoices (with an occasional
// return knocking a bit off), dated over the last year.
func seedBranchSpend(ctx context.Context, store db.Store, branchID, clientID, amountCents int64) {
	if amountCents <= 0 {
		return
	}

	branchClientID := clientID
	if err := store.UpsertClientLink(ctx, db.UpsertClientLinkParams{
		BranchID:       branchID,
		BranchClientID: branchClientID,
		ClientID:       clientID,
	}); err != nil {
		log.Printf("warning: failed to link client %d to branch %d: %v", clientID, branchID, err)
		return
	}

	invoiceCount := rand.Intn(5) + 1
	remaining := amountCents
	for i := 0; i < invoiceCount; i++ {
		share := remaining / int64(invoiceCount-i)
		if share <= 0 {
			continue
		}
		occurredAt := time.Now().AddDate(0, 0, -rand.Intn(365))
		clientRef := fmt.Sprintf("SEED-%d-%d-%d", branchID, clientID, i)

		_, err := store.CreateBranchInvoice(ctx, db.CreateBranchInvoiceParams{
			BranchID:               branchID,
			ClientRef:              clientRef,
			Kind:                   "sales",
			BranchInvoiceCode:      clientRef,
			BranchCashboxAccountID: 1,
			BranchShiftID:          1,
			BranchInventoryID:      1,
			BranchClientID:         nullInt64(branchClientID),
			Discount:               0,
			Subtotal:               share,
			DiscountedTotal:        share,
			GrandTotal:             share,
			LoyaltyPointsDelta:     share / 100,
			OccurredAt:             occurredAt,
		})
		if err != nil {
			log.Printf("warning: failed to create branch invoice for client %d: %v", clientID, err)
			continue
		}
		remaining -= share

		// Occasionally add a small return so not every client's history is
		// a straight line of sales.
		if rand.Intn(6) == 0 && share > 5000 {
			refundCents := -(share / 10)
			returnRef := fmt.Sprintf("SEED-%d-%d-%d-r", branchID, clientID, i)
			_, err := store.CreateBranchInvoice(ctx, db.CreateBranchInvoiceParams{
				BranchID:               branchID,
				ClientRef:              returnRef,
				Kind:                   "return",
				BranchInvoiceCode:      returnRef,
				BranchCashboxAccountID: 1,
				BranchShiftID:          1,
				BranchInventoryID:      1,
				BranchClientID:         nullInt64(branchClientID),
				RelatedClientRef:       sql.NullString{String: clientRef, Valid: true},
				Discount:               0,
				Subtotal:               -refundCents,
				DiscountedTotal:        -refundCents,
				GrandTotal:             refundCents,
				LoyaltyPointsDelta:     refundCents / 100,
				OccurredAt:             occurredAt.Add(time.Hour * 24 * 3),
			})
			if err != nil {
				log.Printf("warning: failed to create return invoice for client %d: %v", clientID, err)
			}
		}
	}
}
