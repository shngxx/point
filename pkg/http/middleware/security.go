package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// Security returns a middleware that sets security headers
func Security() Handler {
	return func(c *fiber.Ctx) error {
		// X-Frame-Options: prevents clickjacking
		c.Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options: prevents MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection: enables XSS filtering
		c.Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: controls referrer information
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy: restricts browser features
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		return c.Next()
	}
}

// SecurityWithCSP returns a middleware that sets security headers including CSP
func SecurityWithCSP(csp string) Handler {
	return func(c *fiber.Ctx) error {
		// Apply basic security headers
		securityHandler := Security()
		if err := securityHandler(c); err != nil {
			return err
		}

		// Content-Security-Policy
		if csp != "" {
			c.Set("Content-Security-Policy", csp)
		}

		return c.Next()
	}
}

