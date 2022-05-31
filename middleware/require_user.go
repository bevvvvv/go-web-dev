package middleware

import (
	"go-web-dev/context"
	"go-web-dev/models"
	"net/http"
	"strings"
)

type UserExists struct {
	models.UserService
}

func (userExists *UserExists) Apply(next http.Handler) http.HandlerFunc {
	return userExists.ApplyFn(next.ServeHTTP)
}

func (userExists *UserExists) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// static assets are not blocked behind login
		path := r.URL.Path
		if strings.HasPrefix(path, "/assets/") || strings.HasPrefix(path, "/images/") {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("remember_token")
		if err != nil {
			next(w, r)
			return
		}
		user, err := userExists.UserService.ByRemember(cookie.Value)
		if err != nil {
			next(w, r)
			return
		}

		// add user data to request context
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)

		next(w, r)
	})
}

// assumes that UserExists has already been run
type UserVerification struct {
	UserExists
}

// assumes that UserExists has already been run
func (userVerification *UserVerification) Apply(next http.Handler) http.HandlerFunc {
	return userVerification.ApplyFn(next.ServeHTTP)
}

// assumes that UserExists has already been run
func (userVerification *UserVerification) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return userVerification.UserExists.ApplyFn(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next(w, r)
	}))
}
