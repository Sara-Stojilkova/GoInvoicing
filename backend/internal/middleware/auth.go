package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend/api"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	ContextUserID   contextKey = "user_id"
	ContextAgencyID contextKey = "agency_id"
	ContextRole     contextKey = "role"
)

func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				api.WriteError(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				api.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				api.WriteError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			sub, err := claims.GetSubject()
			if err != nil {
				api.WriteError(w, http.StatusUnauthorized, "invalid token subject")
				return
			}
			userID, err := uuid.Parse(sub)
			if err != nil {
				api.WriteError(w, http.StatusUnauthorized, "invalid user id in token")
				return
			}

			appMeta, ok := claims["app_metadata"].(map[string]interface{})
			if !ok {
				api.WriteError(w, http.StatusUnauthorized, "missing app_metadata in token")
				return
			}
			agencyIDStr, _ := appMeta["agency_id"].(string)
			agencyID, err := uuid.Parse(agencyIDStr)
			if err != nil {
				api.WriteError(w, http.StatusUnauthorized, "invalid agency_id in token")
				return
			}
			role, _ := appMeta["role"].(string)
			if role == "" {
				api.WriteError(w, http.StatusUnauthorized, "missing role in token")
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, userID)
			ctx = context.WithValue(ctx, ContextAgencyID, agencyID)
			ctx = context.WithValue(ctx, ContextRole, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
