package xfiber

import (
	"strconv"
	"strings"

	commpb "monorepo/proto/xadminpb/commpb"

	"github.com/gofiber/fiber/v2"
)

// ParsePageSortQuery parses common GET list query args:
// page_no/page_size/order_field/order_type, with pn/ps compatibility.
func ParsePageSortQuery(c *fiber.Ctx, defaultPn, defaultPs int32) (*commpb.PageArgs, []*commpb.SortArgs, error) {
	page := &commpb.PageArgs{Pn: defaultPn, Ps: defaultPs}

	pageNoRaw := strings.TrimSpace(c.Query("page_no"))
	if pageNoRaw == "" {
		pageNoRaw = strings.TrimSpace(c.Query("pn"))
	}
	if pageNoRaw != "" {
		pageNo, err := strconv.Atoi(pageNoRaw)
		if err != nil {
			return nil, nil, err
		}
		page.Pn = int32(pageNo)
	}

	pageSizeRaw := strings.TrimSpace(c.Query("page_size"))
	if pageSizeRaw == "" {
		pageSizeRaw = strings.TrimSpace(c.Query("ps"))
	}
	if pageSizeRaw != "" {
		pageSize, err := strconv.Atoi(pageSizeRaw)
		if err != nil {
			return nil, nil, err
		}
		page.Ps = int32(pageSize)
	}

	sort := make([]*commpb.SortArgs, 0, 1)
	if orderField := strings.TrimSpace(c.Query("order_field")); orderField != "" {
		orderType := commpb.OrderType_OT_Desc
		orderTypeRaw := strings.TrimSpace(c.Query("order_type"))
		if strings.EqualFold(orderTypeRaw, "asc") || orderTypeRaw == "1" {
			orderType = commpb.OrderType_OT_Asc
		}
		sort = append(sort, &commpb.SortArgs{
			OrderField: orderField,
			OrderType:  orderType,
		})
	}

	return page, sort, nil
}
