package main

import "testing"

func TestRoutes_HealthCheck(t *testing.T) {
	t.Log("Manual smoke test checklist:")
	t.Log("  1. Start server: JWT_SECRET=x JD_APP_KEY=x JD_APP_SECRET=x JD_SITE_ID=x JD_PID=x go run ./cmd/server/")
	t.Log("  2. GET  /login     → login page renders (200)")
	t.Log("  3. GET  /register  → register page renders (200)")
	t.Log("  4. POST /register  → creates user, redirects to /login (302)")
	t.Log("  5. POST /login     → sets cookies, redirects to / (302)")
	t.Log("  6. GET  /          → convert page renders (200, authenticated)")
	t.Log("  7. POST /convert   → converts link (200, with valid JD creds)")
	t.Log("  8. GET  /orders    → order list renders (200)")
	t.Log("  9. GET  /logout    → clears cookies, redirects to /login (302)")
	t.Log(" 10. Confirm CSRF: POST without X-CSRF-Token → 403")
	t.Log(" 11. Confirm rate limit: POST /convert >10 times/min → 429")
}
