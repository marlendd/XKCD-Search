package middleware

import (
	"net/http"
	"sync"
	"time"
)

func Rate(next http.HandlerFunc, rps int) http.HandlerFunc {
	interval := time.Second / time.Duration(rps)
	last := time.Now()
	var mu sync.Mutex

	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		now := time.Now()
		if last.Before(now) {
			last = now
		}
		next_tick := last.Add(interval)
		last = next_tick
		mu.Unlock()

		if next_tick.After(now) {
			time.Sleep(next_tick.Sub(now))
		}

		next(w, r)
	}
}
