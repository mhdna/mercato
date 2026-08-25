package api

import (
	"testing"
	"time"

	db "github.com/mhdna/kashi/db/sqlc"
)

func TestClampDayHandlesShortMonths(t *testing.T) {
	cases := []struct {
		year  int
		month time.Month
		day   int
		want  int
	}{
		{2026, time.January, 31, 31},
		{2026, time.February, 31, 28}, // 2026 is not a leap year
		{2024, time.February, 31, 29}, // 2024 is a leap year
		{2026, time.April, 31, 30},
		{2026, time.April, 15, 15},
	}
	for _, c := range cases {
		if got := clampDay(c.year, c.month, c.day); got != c.want {
			t.Errorf("clampDay(%d, %s, %d) = %d, want %d", c.year, c.month, c.day, got, c.want)
		}
	}
}

func TestCurrentPeriodBillingCycleStartDay(t *testing.T) {
	// "start every 4th of the month, end every 3rd of the next" -- the
	// exact example from the feature request.
	series := db.BranchTargetSeries{
		StartDay:      4,
		IntervalCount: 1,
		CreatedAt:     time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
	}

	start, end := currentPeriod(series, time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC))
	if !start.Equal(time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("start = %v, want 2026-08-04", start)
	}
	if !end.Equal(time.Date(2026, time.September, 3, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("end = %v, want 2026-09-03", end)
	}

	// Just before the next cycle starts -- still in August's period.
	start, end = currentPeriod(series, time.Date(2026, time.September, 3, 0, 0, 0, 0, time.UTC))
	if !start.Equal(time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("start = %v, want 2026-08-04 (still previous cycle)", start)
	}
	if !end.Equal(time.Date(2026, time.September, 3, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("end = %v, want 2026-09-03", end)
	}

	// On the 4th -- rolled into September's period.
	start, end = currentPeriod(series, time.Date(2026, time.September, 4, 0, 0, 0, 0, time.UTC))
	if !start.Equal(time.Date(2026, time.September, 4, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("start = %v, want 2026-09-04", start)
	}
	if !end.Equal(time.Date(2026, time.October, 3, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("end = %v, want 2026-10-03", end)
	}
}

func TestCurrentPeriodClampsAcrossShortMonths(t *testing.T) {
	// start_day=31 must clamp in every short month it lands in, both for
	// the period start and for computing when the *next* period begins.
	series := db.BranchTargetSeries{
		StartDay:      31,
		IntervalCount: 1,
		CreatedAt:     time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC),
	}

	// February clamps to the 28th (2026 is not a leap year); the period
	// runs Feb 28 through the day before March's clamped start (31st).
	start, end := currentPeriod(series, time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC))
	if !start.Equal(time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("start = %v, want 2026-02-28", start)
	}
	if !end.Equal(time.Date(2026, time.March, 30, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("end = %v, want 2026-03-30", end)
	}
}

func TestCurrentPeriodEveryNMonths(t *testing.T) {
	series := db.BranchTargetSeries{
		StartDay:      1,
		IntervalCount: 3, // quarterly
		CreatedAt:     time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	start, end := currentPeriod(series, time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC))
	if !start.Equal(time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("start = %v, want 2026-07-01", start)
	}
	if !end.Equal(time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("end = %v, want 2026-09-30", end)
	}
}
