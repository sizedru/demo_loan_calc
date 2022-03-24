package rd

import "time"

// ISO формат даты без времени
// OSI обычный формат даты без времени
const (
	ISO = "2006-01-02"
	OSI = "02.01.2006"
)

// Time is date
type Time struct {
	begin, end time.Time
	diffDays   int
}

// SetIntervalDate задать интервал дат
func (t *Time) SetIntervalDate(dBegin, dEnd string) {
	t.begin = Date(dBegin)
	t.end = Date(dEnd)
	t.diffDays = DiffDays(t.begin, t.end)
}

// Date Дата без времени (на утро)
func Date(dStr string) time.Time {
	var ret time.Time
	if len(dStr) >= 10 {
		ret, _ = time.Parse(time.RFC3339, dStr[0:10]+"T00:00:00+03:00")
	} else {
		ret, _ = time.Parse(time.RFC3339, "00-00-0000"+"T00:00:00+03:00")
	}
	return ret
}

// DateMorning Дата без времени (на утро)
func DateMorning(dStr string) time.Time {
	return Date(dStr)
}

// DateEvening Дата без времени (на вечер 23:59:59)
func DateEvening(dStr string) time.Time {
	ret, _ := time.Parse(time.RFC3339, dStr[0:10]+"T23:59:59+03:00")
	return ret
}

// DateTime Дата и время одной строкой
func DateTime(dtStr string) time.Time {
	ret, _ := time.Parse(time.RFC3339, dtStr[0:19]+"+03:00")
	return ret
}

// DateAndTime Дата и время двумя параметрами
func DateAndTime(dStr, tStr string) time.Time {
	ret, _ := time.Parse(time.RFC3339, dStr[0:10]+"T"+tStr[0:8]+"+03:00")
	return ret
}

// DiffDays разница в днях
func DiffDays(a, b time.Time) int {
	if a.After(b) {
		a, b = b, a
	}

	days := -a.YearDay()
	for year := a.Year(); year < b.Year(); year++ {
		days += time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC).YearDay()
	}

	days += b.YearDay()

	return days
}

// AddDays прибавить дни
func AddDays(a time.Time, d int) time.Time {
	return a.AddDate(0, 0, d)
}
