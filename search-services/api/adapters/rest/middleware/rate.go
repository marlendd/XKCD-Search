package middleware

import (
	"net/http"
)

func Rate(next http.HandlerFunc, rps int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
