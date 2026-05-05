package web

import (
	"net/http"
	"strings"
)

const tokenCookieName = "token"

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(tokenCookieName)
		if err != nil || strings.TrimSpace(c.Value) == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
