package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"aidanwoods.dev/go-paseto"
	"github.com/rimdian/rimdian/internal/common/utils"
	"github.com/sirupsen/logrus"
)

func MiddlewarePasetoExtractor(logger *logrus.Logger, issuer string, secretKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// extract token
			authHeader := r.Header.Get("Authorization")
			authParameter := r.URL.Query().Get("rmd_token")

			if authHeader == "" && authParameter == "" {
				next.ServeHTTP(w, r)
				return
			}

			var token string

			if authParameter != "" {
				token = authParameter
			} else {
				authHeaderParts := strings.Fields(authHeader)
				if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
					utils.ReturnJSONError(w, logger, http.StatusNotAcceptable, errors.New("invalid token format"))
					return
				}
				token = authHeaderParts[1]
			}

			// decrypt token
			ctx := r.Context()

			parser := paseto.NewParser()
			parser.AddRule(paseto.ForAudience(TokenAccountAudience))
			parser.AddRule(paseto.IssuedBy(issuer))
			parser.AddRule(paseto.NotExpired())

			// v2 or v4 token
			if strings.HasPrefix(token, "v2") {

				// key, err := paseto.V2SymmetricKeyFromBytes([]byte(secretKey))
				key, err := paseto.V2SymmetricKeyFromBytes([]byte(secretKey))
				if err != nil {
					utils.ReturnJSONError(w, logger, http.StatusUnauthorized, err)
					return
				}

				pasetoToken, err := parser.ParseV2Local(key, token)
				if err != nil {
					utils.ReturnJSONError(w, logger, http.StatusUnauthorized, err)
					return
				}

				ctx = context.WithValue(ctx, AccountTokenContextKey, pasetoToken)

			} else if strings.HasPrefix(token, "v4") {

				key, err := paseto.V4SymmetricKeyFromBytes([]byte(secretKey))
				if err != nil {
					utils.ReturnJSONError(w, logger, http.StatusUnauthorized, err)
					return
				}

				pasetoToken, err := parser.ParseV4Local(key, token, nil)

				if err != nil {
					utils.ReturnJSONError(w, logger, http.StatusUnauthorized, err)
					return
				}

				ctx = context.WithValue(ctx, AccountTokenContextKey, pasetoToken)
			} else {
				utils.ReturnJSONError(w, logger, http.StatusUnauthorized, errors.New("invalid token format"))
				return
			}

			ctx = context.WithValue(ctx, AccountRawTokenContextKey, token)
			// add token context to request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// enforces presence of paseto token
func MiddlewarePasetoRequired(logger *logrus.Logger, secretKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			eventualToken := r.Context().Value(AccountTokenContextKey)

			if eventualToken == nil {
				utils.ReturnJSONError(w, logger, http.StatusUnauthorized, errors.New("an account token is required"))
				return

			}
			next.ServeHTTP(w, r)
		})
	}
}
