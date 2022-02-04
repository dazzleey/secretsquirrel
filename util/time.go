package util

import (
	"time"
)

var daysinMonth = []int{-1, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

func isLeap(year int) bool {
	return year%4 == 0 && (year&100 != 0 || year%400 == 0)
}

func getDaysBeforeMonth() []int {
	var daysBeforeMonth []int
	var dbm int = 0

	for d := range daysinMonth[1:] {
		daysBeforeMonth = append(daysBeforeMonth, dbm)
		dbm += d
	}
	return daysBeforeMonth
}

func daysBeforeYear(year int) float64 {
	y := year - 1

	return float64(y*365) +
		float64(y/4) -
		float64(y/100) +
		float64(y/400)
}

func daysBeforeMonth(year int, month int) int {
	dbm := getDaysBeforeMonth()[month]
	if month > 2 && isLeap(year) {
		dbm += 1
	}
	return dbm
}

// toOrdinal copied from python's datetime.
func ToOrdinal(t time.Time) float64 {
	year, month, day := t.Date()

	dby := daysBeforeYear(year)
	dbm := daysBeforeMonth(year, int(month))

	return dby + float64(dbm) + float64(day)
}
