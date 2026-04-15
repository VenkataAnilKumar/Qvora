package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/labstack/echo/v4"
)

type ClerkClaims struct {
	UserID  string
	OrgID   string
	OrgRole string
}

const (
	clerkClaimsKey = "clerk_claims"
)

// ClerkAuth validates the Clerk JWT in Authorization: Bearer <token>
// Injects ClerkClaims into echo context — does NOT reject unauthenticated requests.
// Use RequireWorkspace() for protected routes.
func ClerkAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if setClaimsFromInternalHeaders(c) {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return next(c)
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			jwksClient := jwks.NewClient(&clerk.ClientConfig{})
			claims, err := jwt.Verify(c.Request().Context(), &jwt.VerifyParams{
				Token:      token,
				JWKSClient: jwksClient,
			})
			if err != nil {
				// Invalid token — continue without claims (RequireWorkspace will reject)
				return next(c)
			}

			c.Set(clerkClaimsKey, &ClerkClaims{
				UserID:  claims.Subject,
				OrgID:   claims.ActiveOrganizationID,
				OrgRole: claims.ActiveOrganizationRole,
			})

			return next(c)
		}
	}
}

// setClaimsFromInternalHeaders enables trusted internal service-to-service auth.
// If INTERNAL_API_KEY is configured, requests must provide matching X-Internal-Api-Key.
func setClaimsFromInternalHeaders(c echo.Context) bool {
	userID := strings.TrimSpace(c.Request().Header.Get("X-User-Id"))
	orgID := strings.TrimSpace(c.Request().Header.Get("X-Org-Id"))
	if userID == "" || orgID == "" {
		return false
	}

	requiredKey := strings.TrimSpace(os.Getenv("INTERNAL_API_KEY"))
	providedKey := strings.TrimSpace(c.Request().Header.Get("X-Internal-Api-Key"))
	if requiredKey != "" {
		if subtle.ConstantTimeCompare([]byte(requiredKey), []byte(providedKey)) != 1 {
			return false
		}
	}

	orgRole := strings.TrimSpace(c.Request().Header.Get("X-Org-Role"))
	if orgRole == "" {
		orgRole = "member"
	}

	c.Set(clerkClaimsKey, &ClerkClaims{
		UserID:  userID,
		OrgID:   orgID,
		OrgRole: orgRole,
	})

	return true
}

// RequireWorkspace rejects requests without a valid workspace context (orgId required).
func RequireWorkspace() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := c.Get(clerkClaimsKey).(*ClerkClaims)
			if !ok || claims == nil || claims.UserID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}
			if claims.OrgID == "" {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "workspace_required"})
			}
			return next(c)
		}
	}
}

// GetClaims retrieves the Clerk claims from echo context.
func GetClaims(c echo.Context) *ClerkClaims {
	claims, _ := c.Get(clerkClaimsKey).(*ClerkClaims)
	return claims
}
