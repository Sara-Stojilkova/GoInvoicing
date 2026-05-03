package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"strings"

	"backend/api"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ContextKey string

const (
	ContextUserID   ContextKey = "user_id"
	ContextAgencyID ContextKey = "agency_id"
	ContextRole     ContextKey = "role"
)

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

// fetchECKeys fetches the JWKS from the Supabase auth endpoint and returns
// a map of key ID → EC public key for ES256 token verification.
func fetchECKeys(supabaseURL string) map[string]*ecdsa.PublicKey {
	keys := map[string]*ecdsa.PublicKey{}
	if supabaseURL == "" {
		return keys
	}
	resp, err := http.Get(supabaseURL + "/auth/v1/.well-known/jwks.json")
	if err != nil {
		log.Printf("middleware: fetch JWKS: %v", err)
		return keys
	}
	defer resp.Body.Close()

	var j jwks
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		log.Printf("middleware: decode JWKS: %v", err)
		return keys
	}

	for _, k := range j.Keys {
		if k.Kty != "EC" || k.Crv != "P-256" {
			continue
		}
		xBytes, err := base64.RawURLEncoding.DecodeString(k.X)
		if err != nil {
			continue
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(k.Y)
		if err != nil {
			continue
		}
		keys[k.Kid] = &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}
	}
	log.Printf("middleware: loaded %d EC key(s) from Supabase JWKS", len(keys))
	return keys
}

// Authenticate validates JWTs and injects userID, agencyID, and role into the
// request context. Supports both HS256 (legacy secret) and ES256 (Supabase JWKS).
func Authenticate(jwtSecret, supabaseURL string) func(http.Handler) http.Handler {
	ecKeys := fetchECKeys(supabaseURL)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				api.WriteError(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				switch t.Method.(type) {
				case *jwt.SigningMethodHMAC:
					return []byte(jwtSecret), nil
				case *jwt.SigningMethodECDSA:
					kid, _ := t.Header["kid"].(string)
					key, ok := ecKeys[kid]
					if !ok {
						return nil, jwt.ErrSignatureInvalid
					}
					return key, nil
				default:
					return nil, jwt.ErrSignatureInvalid
				}
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
			agencyIDStr, ok := appMeta["agency_id"].(string)
			if !ok {
				api.WriteError(w, http.StatusUnauthorized, "invalid agency_id in token")
				return
			}
			agencyID, err := uuid.Parse(agencyIDStr)
			if err != nil {
				api.WriteError(w, http.StatusUnauthorized, "invalid agency_id in token")
				return
			}
			role, ok := appMeta["role"].(string)
			if !ok || role == "" {
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
