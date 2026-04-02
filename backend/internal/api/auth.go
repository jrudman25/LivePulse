package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
)

// SetClerkKey initializes the Clerk SDK with the secret key
func SetClerkKey(secret string) {
	clerk.SetKey(secret)
}

// ClerkMiddleware verifies the Clerk JWT on incoming HTTP requests.
func ClerkMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionToken := r.Header.Get("Authorization")
		sessionToken = strings.TrimPrefix(sessionToken, "Bearer ")

		if sessionToken == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{
			Token: sessionToken,
		})
		if err != nil {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		// Inject UserID into request context
		ctx := context.WithValue(r.Context(), "user_id", claims.Subject)
		next(w, r.WithContext(ctx))
	}
}

// VerifyTokenManually verifies a Clerk token passed via URL query or payload
func VerifyTokenManually(ctx context.Context, token string) (string, error) {
	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{
		Token: token,
	})
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}
