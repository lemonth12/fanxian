package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateBucket struct {
	timestamps []time.Time
}

var (
	rateMap  = make(map[string]*rateBucket)
	rateLock sync.Mutex
)

func RateLimitMiddleware(perMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if userID, exists := c.Get("userID"); exists {
			if uid, ok := userID.(uint); ok && uid != 0 {
				key = "user_" + fmt.Sprintf("%d", uid)
			}
		}

		rateLock.Lock()
		bucket, ok := rateMap[key]
		if !ok {
			bucket = &rateBucket{}
			rateMap[key] = bucket
		}
		rateLock.Unlock()

		bucket.timestamps = append(bucket.timestamps, time.Now())

		cutoff := time.Now().Add(-1 * time.Minute)
		valid := make([]time.Time, 0)
		for _, ts := range bucket.timestamps {
			if ts.After(cutoff) {
				valid = append(valid, ts)
			}
		}
		bucket.timestamps = valid

		if len(valid) > perMinute {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "操作太频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}
