package date

import "time"

const DATE_FORMAT string = "2006-01-02"

func GetDateFromString(dateString string) time.Time {
	t, err := time.Parse(DATE_FORMAT, dateString)
	if err != nil {
		return time.Time{}
	}
	return t
}

func GetDateAsString(i interface{}) string {
	if i == nil {
		return ""
	} else {
		return i.(time.Time).Format(DATE_FORMAT)
	}
}

func FormatDate(t time.Time) string {
	return t.Format(DATE_FORMAT)
}
