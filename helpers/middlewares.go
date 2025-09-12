package helpers

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Visitor stores limiter and block info
type Visitor struct {
	Limiter      *rate.Limiter
	BlockedUntil time.Time
}

// Thread-safe map of IPs to Visitor
var visitors = make(map[string]*Visitor)
var mu sync.Mutex

// getVisitor returns a limiter for an IP, creating it if needed
func getVisitor(ip string, rps float64, burst int) *Visitor {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		v = &Visitor{
			Limiter:      rate.NewLimiter(rate.Limit(rps), burst),
			BlockedUntil: time.Time{},
		}
		visitors[ip] = v
	}

	return v
}

// RateLimitMiddleware applies rate limiting per IP for specified routes
// rps: requests per second
// burst: max burst per IP
// blockDuration: time to block IP after limit exceeded
// routes: paths to apply limiter to
func RateLimitMiddleware(rps float64, burst int, blockDuration time.Duration, routes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		for _, route := range routes {
			if strings.HasPrefix(path, route) {
				ip := c.ClientIP()
				v := getVisitor(ip, rps, burst)

				mu.Lock()
				// Check if IP is currently blocked
				if !v.BlockedUntil.IsZero() && time.Now().Before(v.BlockedUntil) {
					blockedFor := v.BlockedUntil.Sub(time.Now())
					mu.Unlock()
					log.Printf("IP %s blocked for %v on %s", ip, blockedFor, path)
					c.JSON(http.StatusTooManyRequests, gin.H{
						"success":       false,
						"message":       "IP blocked",
						"ip":            ip,
						"path":          path,
						"blocked_until": v.BlockedUntil.Format(time.RFC3339),
					})
					c.Abort()
					return
				}

				// Try allowing request
				if !v.Limiter.Allow() {
					v.BlockedUntil = time.Now().Add(blockDuration)
					mu.Unlock()
					log.Printf("Rate limit exceeded for IP %s on %s. Blocked until %v", ip, path, v.BlockedUntil)
					c.JSON(http.StatusTooManyRequests, gin.H{
						"success":       false,
						"message":       "Rate limit exceeded, IP temporarily blocked",
						"ip":            ip,
						"path":          path,
						"blocked_until": v.BlockedUntil.Format(time.RFC3339),
					})
					c.Abort()
					return
				}
				mu.Unlock()
				break
			}
		}
		c.Next()
	}
}

// StartCleanup periodically removes old visitors to free memory
func StartCleanup(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			mu.Lock()
			for ip, v := range visitors {
				// Remove visitors whose limiter is idle and not blocked
				if v.Limiter.AllowN(time.Now(), 0) && time.Now().After(v.BlockedUntil) {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// CORSMiddleware returns a CORS middleware for allowed origins
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
