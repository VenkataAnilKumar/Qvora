package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/db"
	"github.com/qvora/api/internal/store"
)

const rlsScopeKey = "rls_scope"

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
	if requiredKey == "" {
		// If INTERNAL_API_KEY is not configured, disable internal header auth entirely.
		return false
	}
	providedKey := strings.TrimSpace(c.Request().Header.Get("X-Internal-Api-Key"))
	if subtle.ConstantTimeCompare([]byte(requiredKey), []byte(providedKey)) != 1 {
		return false
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
			// Acquire a dedicated connection with app.org_id set, so RLS policies
			// (which key on current_setting('app.org_id', TRUE)) actually have
			// something to check — see store.AcquireForOrg for why this can't be
			// done on the shared pool-wrapped Queries(). Falls through without a
			// scope if acquisition fails so routes keep working against the
			// pre-existing pool-backed behavior rather than hard-failing every
			// request on this still-being-proven-out code path.
			scope, err := store.AcquireForOrg(c.Request().Context(), claims.OrgID)
			if err == nil {
				defer scope.Close(c.Request().Context())
				c.Set(rlsScopeKey, scope)
			}
			return next(c)
		}
	}
}

// GetScopedQueries returns the request-scoped *db.Queries set by
// RequireWorkspace, with RLS actually enforced (a dedicated connection with
// app.org_id set). Falls back to the shared pool-backed store.Queries — which,
// under the `postgres` superuser DATABASE_URL still configured today, bypasses
// RLS entirely — if no scope is present (e.g. internal/webhook routes that
// don't go through RequireWorkspace, or if scope acquisition failed).
// Domain handlers should prefer this over calling store.Queries directly
// wherever the request is workspace-scoped.
func GetScopedQueries(c echo.Context) (*db.Queries, error) {
	if scope, ok := c.Get(rlsScopeKey).(*store.RequestScope); ok && scope != nil {
		return scope.Queries, nil
	}
	return store.Queries(c.Request().Context())
}

// GetScopedConn is GetScopedQueries' counterpart for call sites using raw SQL
// via store.Pool() directly rather than sqlc's generated Queries wrapper.
// Returns db.DBTX (Exec/Query/QueryRow) so call sites change only the source
// of their handle, not how they use it — a drop-in replacement for
// store.Pool() at call sites that previously used it directly.
func GetScopedConn(c echo.Context) (db.DBTX, error) {
	if scope, ok := c.Get(rlsScopeKey).(*store.RequestScope); ok && scope != nil {
		return scope.Conn(), nil
	}
	if store.Pool() == nil {
		return nil, errors.New("database_not_initialized")
	}
	return store.Pool(), nil
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
