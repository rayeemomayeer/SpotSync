package platform

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	reservationLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "reservation_latency_seconds",
		Help:    "Reservation handler latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "status"})

	oversellRejected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oversell_attempts_rejected_total",
		Help: "Total reservation attempts rejected because zone was full",
	})

	zoneAvailCacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zone_availability_cache_hits_total",
		Help: "Zone availability cache hits on GET /zones/:id",
	})

	zoneAvailCacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zone_availability_cache_misses_total",
		Help: "Zone availability cache misses on GET /zones/:id",
	})

	outboxDeadLettered = promauto.NewCounter(prometheus.CounterOpts{
		Name: "outbox_dead_lettered_total",
		Help: "Outbox events moved to dead-letter queue",
	})
)

func MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			status := strconv.Itoa(c.Response().Status)
			path := c.Path()
			if path == "/api/v1/reservations" && c.Request().Method == "POST" {
				reservationLatency.WithLabelValues("POST", status).Observe(time.Since(start).Seconds())
				if c.Response().Status == 409 {
					oversellRejected.Inc()
				}
			}
			return err
		}
	}
}

func RecordOversellRejected() {
	oversellRejected.Inc()
}

func RecordZoneAvailCacheHit() {
	zoneAvailCacheHits.Inc()
}

func RecordZoneAvailCacheMiss() {
	zoneAvailCacheMisses.Inc()
}

func RecordOutboxDeadLetter() {
	outboxDeadLettered.Inc()
}
