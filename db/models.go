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

type PeriodLogRow struct {
	CategoryName string
	HabitName    string
	Date         time.Time
	Count        int
}
