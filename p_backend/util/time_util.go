package util

import (
	"errors"
	"monorepo/proto/xadminpb/commpb"
	"time"
)

func CheckTimeInRange(t time.Time, timeRange *commpb.TimeRange) (bool, error) {
	stStr, etStr := timeRange.StartDt, timeRange.EndDt
	st, e1 := time.ParseInLocation(time.DateTime, stStr, time.Local)
	et, e2 := time.ParseInLocation(time.DateTime, etStr, time.Local)
	if e1 != nil || e2 != nil {
		return false, errors.New("invalid time range")
	}
	return t.After(st) && t.Before(et), nil
}

func Time2Unix(t *time.Time) int32 {
	if t == nil {
		return 0
	}
	return int32(t.Unix())
}
