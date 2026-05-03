package middleware

import (
	"errors"
	"net/http"
	"strings"

	"yadro.com/course/api/core"
)
//go:generate mockgen -source=auth.go -destination=../../../mocks/mock_tokenVerifier.go -package=mocks

type TokenVerifier interface {
	Verify(token string) error
}

func Auth(next http.HandlerFunc, verifier TokenVerifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token := strings.TrimPrefix(header, "Token ")
		
		if header == token || token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		
		if err := verifier.Verify(token); err != nil {
			if errors.Is(err, core.ErrNotAuthorized) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, "unknown error", http.StatusInternalServerError)
			return
		}

		next(w, r)
	}
}
