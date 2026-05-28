package util

import (
	"monorepo/pkg/xerr"
	"monorepo/proto/xadminpb/commpb"
	"time"
)

func NormalizePageArgs(page *commpb.PageArgs) (pn, ps, offset int) {
	return int(page.Pn), int(page.Ps), int((page.Pn - 1) * page.Ps)
}

func CheckTimeRange(r *commpb.TimeRange) error {
	stStr, etStr := r.StartDt, r.EndDt
	st, e1 := time.ParseInLocation(time.DateTime, stStr, time.Local)
	et, e2 := time.ParseInLocation(time.DateTime, etStr, time.Local)
	if e1 != nil || e2 != nil {
		return xerr.NewWithDetail(xerr.CodeParamError, "invalid event time")
	}
	if st.After(et) {
		return xerr.NewWithDetail(xerr.CodeParamError, "invalid event time range")
	}
	return nil
}
