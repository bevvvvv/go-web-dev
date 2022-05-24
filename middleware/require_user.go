package middleware

import (
	"fmt"
	"go-web-dev/models"
	"net/http"
)

type UserVerification struct {
	models.UserService
}

func (uVerification *UserVerification) Apply(next http.Handler) http.HandlerFunc {
	return uVerification.ApplyFn(next.ServeHTTP)
}

func (uVerification *UserVerification) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		user, err := uVerification.UserService.ByRemember(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
		fmt.Println("Serving page to:", user)

		next(w, r)
	})
}
