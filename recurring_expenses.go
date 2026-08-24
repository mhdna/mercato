package main

import (
	"context"
	"log"
	"time"

	db "github.com/mhdna/kashi/db/sqlc"
)

// recurringExpenseCheckInterval is how often the scheduler looks for due
// recurring expenses. Templates fire on day/month granularity, so checking
// once an hour is more than fine -- it just needs to run at least once
// within the smallest supported interval (a day).
const recurringExpenseCheckInterval = time.Hour

// runRecurringExpenseScheduler processes due recurring expenses immediately
// on startup (so a template that came due while the server was down still
// fires promptly) and then on a fixed tick, for the life of the process.
func runRecurringExpenseScheduler(ctx context.Context, store db.Store) {
	fireDueRecurringExpenses(ctx, store)

	ticker := time.NewTicker(recurringExpenseCheckInterval)
	defer ticker.Stop()
	for range ticker.C {
		fireDueRecurringExpenses(ctx, store)
	}
}

func fireDueRecurringExpenses(ctx context.Context, store db.Store) {
	due, err := store.ListDueRecurringExpenses(ctx, time.Now())
	if err != nil {
		log.Printf("recurring expenses: list due: %v", err)
		return
	}

	for _, recurring := range due {
		if _, err := store.FireRecurringExpenseTx(ctx, recurring); err != nil {
			log.Printf("recurring expense %d: fire: %v", recurring.ID, err)
		}
	}
}
