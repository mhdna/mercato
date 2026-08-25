package api

import (
	"context"
	"database/sql"
	"log"
	"time"

	db "github.com/mhdna/kashi/db/sqlc"
)

// branchTargetSeriesCheckInterval mirrors recurringExpenseCheckInterval
// (see recurring_expenses.go) -- periods fire on month granularity at the
// coarsest, so checking once an hour is more than fine.
const branchTargetSeriesCheckInterval = time.Hour

// RunBranchTargetSeriesScheduler ensures every active recurring target
// series has a branch_targets row for its current period, immediately on
// startup (so a period that started while the server was down still gets
// created promptly) and then on a fixed tick, for the life of the process.
func RunBranchTargetSeriesScheduler(ctx context.Context, store db.Store) {
	fireDueBranchTargetSeries(ctx, store)

	ticker := time.NewTicker(branchTargetSeriesCheckInterval)
	defer ticker.Stop()
	for range ticker.C {
		fireDueBranchTargetSeries(ctx, store)
	}
}

func fireDueBranchTargetSeries(ctx context.Context, store db.Store) {
	series, err := store.ListActiveBranchTargetSeries(ctx)
	if err != nil {
		log.Printf("branch target series: list active: %v", err)
		return
	}

	now := time.Now().UTC()
	for _, s := range series {
		start, end := currentPeriod(s, now)
		seriesID := sql.NullInt64{Int64: s.ID, Valid: true}

		_, err := store.GetBranchTargetBySeriesAndStart(ctx, db.GetBranchTargetBySeriesAndStartParams{
			SeriesID: seriesID,
			DateFrom: start,
		})
		if err == nil {
			continue // this period's row already exists
		}

		if _, err := store.CreateGeneratedBranchTarget(ctx, db.CreateGeneratedBranchTargetParams{
			BranchID:     s.BranchID,
			DateFrom:     start,
			DateTo:       end,
			TargetAmount: s.TargetAmount,
			Color:        s.Color,
			SeriesID:     seriesID,
		}); err != nil {
			log.Printf("branch target series %d: create period %s: %v", s.ID, start.Format("2006-01-02"), err)
		}
	}
}

// clampDay returns startDay clamped to the last day of the given month --
// e.g. startDay=31 in February becomes 28 (or 29). This is what makes
// periods handle months with fewer or more days correctly, instead of
// Go's time.AddDate silently overflowing a day-of-month into the next
// month (Jan 31 + 1 month = Mar 3, not Feb 28).
func clampDay(year int, month time.Month, startDay int) int {
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if startDay > lastDay {
		return lastDay
	}
	return startDay
}

// periodStart returns the period-start date landing in the given
// year/month for a series with the given startDay (clamped to that
// month's length).
func periodStart(year int, month time.Month, startDay int) time.Time {
	return time.Date(year, month, clampDay(year, month, startDay), 0, 0, 0, 0, time.UTC)
}

// addIntervals returns the year/month intervalCount*n months after the
// given one, normalizing month overflow (e.g. month=12, n=1 -> next year,
// month=1).
func addIntervals(year int, month time.Month, months int) (int, time.Month) {
	total := int(month) - 1 + months
	year += total / 12
	month = time.Month(total%12) + 1
	if total%12 < 0 {
		year--
		month += 12
	}
	return year, month
}

// currentPeriod returns the [start, end] date range (inclusive) of the
// period a series is currently in, as of `now`. Starts walking from the
// series' created_at month and advances by whole intervals until the next
// period's start would be after `now` -- series run for a bounded number
// of months in practice, so this loop is cheap and, unlike tracking a
// separately-advanced "next due" pointer, is naturally self-healing after
// any amount of downtime.
func currentPeriod(s db.BranchTargetSeries, now time.Time) (start, end time.Time) {
	year, month, _ := s.CreatedAt.Date()
	start = periodStart(year, month, int(s.StartDay))

	for {
		nextYear, nextMonth := addIntervals(year, month, int(s.IntervalCount))
		nextStart := periodStart(nextYear, nextMonth, int(s.StartDay))
		if nextStart.After(now) {
			end = nextStart.AddDate(0, 0, -1)
			return start, end
		}
		year, month, start = nextYear, nextMonth, nextStart
	}
}
