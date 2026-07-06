package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

var (
	paginationExposeHeaders = []string{"X-Total-Count", "X-Page", "X-Limit"}
	defaultAllowHeaders     = []string{
		echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept,
		echo.HeaderAuthorization, echo.HeaderXRequestID, "X-Demo-Reservation",
	}
)

func CORS(allowedOrigins []string) echo.MiddlewareFunc {
	if len(allowedOrigins) == 0 {
		return echomw.CORSWithConfig(echomw.CORSConfig{
			AllowOrigins:   []string{"*"},
			AllowMethods:   []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
			AllowHeaders:   defaultAllowHeaders,
			ExposeHeaders:  paginationExposeHeaders,
		})
	}

	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:   allowedOrigins,
		AllowMethods:   []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
		AllowHeaders:   defaultAllowHeaders,
		ExposeHeaders:  paginationExposeHeaders,
	})
}
