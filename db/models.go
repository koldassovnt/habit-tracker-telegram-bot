package db

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	// ErrNoLog means the habit exists but has no log on the requested day.
	ErrNoLog = errors.New("no log")
)

type Category struct {
	ID     int64
	Name   string
	UserID int64
}

type Habit struct {
	ID         int64
	CategoryID int64
	Name       string
}

type StatusRow struct {
	CategoryName string
	HabitName    string
	Count        int
}

type HabitReminder struct {
	HabitName string
	Hour      int
}

// DueReminder is a habit whose reminder hour has arrived for its owner.
type DueReminder struct {
	UserID       int64
	Habit        Habit
	TrackedToday bool
}

type PeriodLogRow struct {
	CategoryName string
	HabitName    string
	Date         time.Time
	Count        int
}
