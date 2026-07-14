package middleware

import (
	"flower-lottery-backend/config"
	tokenjwt "flower-lottery-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newAuthMiddlewareTestManager() *tokenjwt.Manager {
	return tokenjwt.New(config.JWT{
		Issuer:              "flower-lottery-test",
		Secret:              "middleware-test-secret",
		AccessExpireMinutes: 15,
		RefreshExpireHours:  24,
	})
}

func issueTestAccessToken(t *testing.T, manager *tokenjwt.Manager, subjectType string) string {
	t.Helper()
	pair, _, err := manager.Issue(42, subjectType)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	return pair.AccessToken
}

func runAdminAuthRequest(manager *tokenjwt.Manager, authorization string) (int, uint64) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var adminID uint64
	router.GET("/admin", AdminAuth(manager), func(c *gin.Context) {
		adminID = CurrentAdminID(c)
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.Header.Set("Authorization", authorization)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response.Code, adminID
}

func TestAdminAuthAcceptsOnlyBearerAdminTokens(t *testing.T) {
	manager := newAuthMiddlewareTestManager()
	adminToken := issueTestAccessToken(t, manager, "admin")
	userToken := issueTestAccessToken(t, manager, "user")

	status, adminID := runAdminAuthRequest(manager, "Bearer "+adminToken)
	if status != http.StatusNoContent || adminID != 42 {
		t.Fatalf("valid admin token returned status=%d adminID=%d", status, adminID)
	}

	status, _ = runAdminAuthRequest(manager, "Bearer "+userToken)
	if status != http.StatusUnauthorized {
		t.Fatalf("user token returned status=%d, want %d", status, http.StatusUnauthorized)
	}

	status, _ = runAdminAuthRequest(manager, "Basic "+adminToken)
	if status != http.StatusUnauthorized {
		t.Fatalf("non-Bearer token returned status=%d, want %d", status, http.StatusUnauthorized)
	}
}
