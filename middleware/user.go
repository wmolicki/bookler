package middleware

import (
	"net/http"

	"github.com/wmolicki/bookler/context"
	"github.com/wmolicki/bookler/models"
)

func NewUserMiddleware(us *models.UserService) *UserMiddleware {
	return &UserMiddleware{us: us}
}

type UserMiddleware struct {
	us *models.UserService
}

func (um *UserMiddleware) AddUser(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(models.AuthCookieName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		token := cookie.Value
		user, err := um.us.ByRememberToken(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
