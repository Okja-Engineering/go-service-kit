package auth

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

type JWTValidator struct {
	clientID string
	scope    string
	jwks     *keyfunc.JWKS
}

type PassthroughValidator struct {
}

type Validator interface {
	Middleware(next http.Handler) http.Handler
	Protect(next http.HandlerFunc) http.HandlerFunc
}

func NewJWTValidator(clientID string, jwksURL string, scope string) JWTValidator {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Duration(1) * time.Hour,
	})

	if err != nil {
		log.Printf("### 🔐 Auth: Failed to fetch the JWKS. Error: %s", err)
	} else {
		log.Printf("### 🔐 Auth: Enabling auth, JWKS fetched from %s", jwksURL)
	}

	return JWTValidator{
		clientID: clientID,
		scope:    scope,
		jwks:     jwks,
	}
}

func NewPassthroughValidator() PassthroughValidator {
	return PassthroughValidator{}
}

func (v JWTValidator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validateRequest(r, v.clientID, v.scope, v.jwks) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (v JWTValidator) Protect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateRequest(r, v.clientID, v.scope, v.jwks) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (v PassthroughValidator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (v PassthroughValidator) Protect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}
}

func validateRequest(r *http.Request, clientID string, scope string, jwks *keyfunc.JWKS) bool {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) == 0 {
		return false
	}

	authParts := strings.Split(authHeader, " ")
	if len(authParts) != 2 {
		return false
	}

	if strings.ToLower(authParts[0]) != "bearer" {
		return false
	}

	if jwks == nil {
		log.Printf("### 🔐 Auth: No JWKS, cannot validate token, denying access")
		return false
	}

	token, err := jwt.Parse(authParts[1], jwks.Keyfunc)
	if err != nil {
		log.Printf("### 🔐 Auth: Failed to parse the JWT. Error: %s", err)
		return false
	}

	claims := token.Claims.(jwt.MapClaims)

	if !strings.Contains(claims["scp"].(string), scope) {
		log.Printf("### 🔐 Auth: Scope '%s' is missing from token scope '%s'", scope, claims["scp"])
		return false
	}

	audience := claims["aud"]
	if strings.HasPrefix(audience.(string), "api://") {
		audience = strings.TrimPrefix(audience.(string), "api://")
	}

	if audience != clientID {
		log.Printf("### 🔐 Auth: Token audience '%s' does not match '%s'", claims["aud"], clientID)
		return false
	}

	return true
}
