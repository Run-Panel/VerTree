package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (adjust based on your needs)
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none';"
		c.Header("Content-Security-Policy", csp)

		// Strict Transport Security (only for HTTPS)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// XSSProtection provides additional XSS protection by sanitizing input
func XSSProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for suspicious patterns in query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsXSS(value) {
					c.JSON(http.StatusBadRequest, models.BadRequestResponse(
						"Potential XSS detected in query parameter: "+key,
						nil,
					))
					c.Abort()
					return
				}
			}
		}

		// Check User-Agent for suspicious patterns
		userAgent := c.GetHeader("User-Agent")
		if containsXSS(userAgent) {
			c.JSON(http.StatusBadRequest, models.BadRequestResponse(
				"Potential XSS detected in User-Agent",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// containsXSS checks if a string contains potential XSS patterns
func containsXSS(input string) bool {
	if input == "" {
		return false
	}

	// Convert to lowercase for case-insensitive matching
	lower := strings.ToLower(input)

	// XSS patterns to check for
	xssPatterns := []string{
		"<script",
		"</script>",
		"javascript:",
		"vbscript:",
		"onload=",
		"onerror=",
		"onclick=",
		"onmouseover=",
		"onfocus=",
		"onblur=",
		"onchange=",
		"onsubmit=",
		"<iframe",
		"<object",
		"<embed",
		"<form",
		"<meta",
		"<link",
		"<style",
		"</style>",
		"expression(",
		"url(",
		"@import",
		"<!--",
		"-->",
	}

	for _, pattern := range xssPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	// Check for HTML entity encoding attempts
	if strings.Contains(input, "&#") || strings.Contains(input, "&lt;") || strings.Contains(input, "&gt;") {
		return true
	}

	return false
}

// SQLInjectionProtection provides basic SQL injection protection
func SQLInjectionProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check query parameters for SQL injection patterns
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsSQLInjection(value) {
					c.JSON(http.StatusBadRequest, models.BadRequestResponse(
						"Potential SQL injection detected in query parameter: "+key,
						nil,
					))
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// containsSQLInjection checks if a string contains potential SQL injection patterns
func containsSQLInjection(input string) bool {
	if input == "" {
		return false
	}

	// Convert to lowercase for case-insensitive matching
	lower := strings.ToLower(input)

	// SQL injection patterns
	sqlPatterns := []string{
		"'",
		"\"",
		";",
		"--",
		"/*",
		"*/",
		"union",
		"select",
		"insert",
		"update",
		"delete",
		"drop",
		"create",
		"alter",
		"exec",
		"execute",
		"sp_",
		"xp_",
	}

	// Use regex for more sophisticated detection
	suspiciousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`'\s*or\s*'1'\s*=\s*'1`), // '1'='1'
		regexp.MustCompile(`"\s*or\s*"1"\s*=\s*"1`), // "1"="1"
		regexp.MustCompile(`'\s*or\s*1\s*=\s*1`),    // '1=1
		regexp.MustCompile(`"\s*or\s*1\s*=\s*1`),    // "1=1
		regexp.MustCompile(`union\s+select`),        // union select
		regexp.MustCompile(`0x[0-9a-f]+`),           // hex values
		regexp.MustCompile(`char\(\d+\)`),           // char() function
		regexp.MustCompile(`concat\s*\(`),           // concat function
		regexp.MustCompile(`substring\s*\(`),        // substring function
	}

	for _, pattern := range sqlPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	for _, regex := range suspiciousPatterns {
		if regex.MatchString(lower) {
			return true
		}
	}

	return false
}

// NoCache adds no-cache headers to prevent caching of sensitive data
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// IPWhitelist creates a middleware that only allows specific IP addresses
func IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	allowedIPMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		allowedIPMap[ip] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if len(allowedIPs) > 0 && !allowedIPMap[clientIP] {
			c.JSON(http.StatusForbidden, models.ErrorResponseWithCodeAndError(
				http.StatusForbidden,
				"IP address not allowed",
				"IP_NOT_ALLOWED",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
