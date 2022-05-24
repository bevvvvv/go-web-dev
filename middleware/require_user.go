package middleware

import (
	"go-web-dev/context"
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
		// add user data to request context
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)

		next(w, r)
	})
}
