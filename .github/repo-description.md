SpotSync is a high-concurrency reservation engine built in Go for managing finite parking and EV charging resources. The core challenge it solves: preventing overbooking when multiple drivers race for the last available spot under high contention.

Key Highlights:
- Never oversells zones: Uses database transactions with SELECT … FOR UPDATE row locking to serialize concurrent reservations per zone, ensuring the capacity invariant is maintained.
- Complete parking management: Supports zone creation with configurable types (general, EV charging, covered), pricing, and capacity; dynamic availability calculation; and driver-scoped reservations.
- Production-ready architecture: Clean layered design (handlers → services → repositories), JWT + bcrypt authentication with role-based access control (driver/admin), comprehensive validation, and structured logging.
- Fully tested: Unit tests, race detector, integration tests with testcontainers, and contract/API replay tests including a "stampede test" that fires 50 concurrent reservations at a 1-capacity zone.
- Cloud-native: Deployed on Render with a Neon PostgreSQL backend; includes Docker, migrations, and health/readiness endpoints.

Tech Stack: Go 1.25+ | Echo v4 | GORM | PostgreSQL | JWT + bcrypt | golang-migrate

SpotSync is ideal for learning production Go patterns or building a reliable parking/charging reservation system from scratch.
