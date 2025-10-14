package service

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// visitor wraps a rate limiter with its last-seen timestamp.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var visitors = make(map[string]*visitor)
var mu sync.Mutex

// getVisitor returns or creates a rate limiter for an IP.
func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		// 30 requests per minute → 0.5 RPS, burst 5
		limiter := rate.NewLimiter(rate.Limit(0.5), 5)
		visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	// update last-seen timestamp
	v.lastSeen = time.Now()
	return v.limiter
}

// background goroutine to remove old entries
func CleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 5*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// Gin middleware
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getVisitor(ip)

		if limiter.Allow() {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "Too many requests — limit is 30 per minute.",
			})
		}
	}
}
