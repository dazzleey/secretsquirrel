package util

import (
	"fmt"
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

func TimeStr(dur time.Duration) (res string) {

	sec := int(dur.Seconds())
	wks, sec := sec/604800, sec%604800
	ds, sec := sec/86400, sec%86400
	hrs, sec := sec/3600, sec%3600
	mins, sec := sec/60, sec%60
	CommaRequired := false

	if wks != 0 {
		res += fmt.Sprintf("%dw", wks)
		CommaRequired = true
	}
	if ds != 0 {
		if CommaRequired {
			res += ", "
		}
		res += fmt.Sprintf("%dd", ds)
		CommaRequired = true
	}
	if hrs != 0 {
		if CommaRequired {
			res += ", "
		}
		res += fmt.Sprintf("%dh", hrs)
		CommaRequired = true
	}
	if mins != 0 {
		if CommaRequired {
			res += ", "
		}
		res += fmt.Sprintf("%dm", mins)
		CommaRequired = true
	}

	if sec != 0 {
		if CommaRequired {
			res += ", "
		}
		res += fmt.Sprintf("%ds", sec)
	}

	return res
}
