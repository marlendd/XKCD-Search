package middleware

import (
	"net/http"
)

func Concurrency(next http.HandlerFunc, limit int) http.HandlerFunc {
	limitChan := make(chan struct{}, limit)

	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case limitChan <- struct{}{}:
			defer func(){ <-limitChan }()
			next(w, r)
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}
