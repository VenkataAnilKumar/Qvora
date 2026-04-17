package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"slices"
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

// RequireWriteAccess enforces read-only access for viewers.
// Member and admin roles are allowed to perform mutating actions.
func RequireWriteAccess() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := GetClaims(c)
			if claims == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			role := normalizeOrgRole(claims.OrgRole)
			if role == "viewer" {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
			}

			return next(c)
		}
	}
}

// RequireAdmin enforces admin-only permissions for membership and billing-critical actions.
func RequireAdmin() echo.MiddlewareFunc {
	return requireAnyRole("admin")
}

func requireAnyRole(allowed ...string) echo.MiddlewareFunc {
	normalizedAllowed := make([]string, 0, len(allowed))
	for _, role := range allowed {
		normalizedAllowed = append(normalizedAllowed, normalizeOrgRole(role))
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := GetClaims(c)
			if claims == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			role := normalizeOrgRole(claims.OrgRole)
			if !slices.Contains(normalizedAllowed, role) {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
			}

			return next(c)
		}
	}
}

func normalizeOrgRole(role string) string {
	normalized := strings.ToLower(strings.TrimSpace(role))
	if normalized == "" {
		return "member"
	}

	if strings.Contains(normalized, ":") {
		parts := strings.Split(normalized, ":")
		normalized = parts[len(parts)-1]
	}

	switch normalized {
	case "admin", "member", "viewer":
		return normalized
	default:
		return "member"
	}
}

// GetClaims retrieves the Clerk claims from echo context.
func GetClaims(c echo.Context) *ClerkClaims {
	claims, _ := c.Get(clerkClaimsKey).(*ClerkClaims)
	return claims
}
