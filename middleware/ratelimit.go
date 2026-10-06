package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"uts-044/helper"
)

type loginAttempt struct {
	count    int
	resetAt  time.Time
}

var (
	attempts = make(map[string]*loginAttempt)
	mu       sync.Mutex
)

func RateLimit(maxAttempts int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		mu.Lock()
		defer mu.Unlock()

		now := time.Now()
		a, exists := attempts[ip]
		if !exists || now.After(a.resetAt) {
			attempts[ip] = &loginAttempt{count: 1, resetAt: now.Add(window)}
			return c.Next()
		}

		if a.count >= maxAttempts {
			return helper.Fail(c, fiber.StatusTooManyRequests, "Terlalu banyak percobaan login, coba lagi nanti")
		}

		a.count++
		return c.Next()
	}
}
