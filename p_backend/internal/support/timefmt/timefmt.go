package timefmt

import "time"

func RFC3339(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

func RFC3339Ptr(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

func RFC3339Unix(seconds int64) string {
	if seconds <= 0 {
		return ""
	}
	return time.Unix(seconds, 0).Format(time.RFC3339)
}

func DateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.In(time.Local).Format(time.DateTime)
}
