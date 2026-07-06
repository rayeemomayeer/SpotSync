package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	HeaderTotalCount = "X-Total-Count"
	HeaderPage       = "X-Page"
	HeaderLimit      = "X-Limit"
)

func setPaginationHeaders(c echo.Context, total int64, page, limit int) {
	c.Response().Header().Set(HeaderTotalCount, strconv.FormatInt(total, 10))
	c.Response().Header().Set(HeaderPage, strconv.Itoa(page))
	c.Response().Header().Set(HeaderLimit, strconv.Itoa(limit))
}
