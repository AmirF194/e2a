package httpapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/tokencanopy/e2a/internal/identity"
)

// Pins chi's current no-StrictSlash behavior: a collapsed ".." arrives here
// as a trailing-slash request, and it must not route to the parent's
// DELETE handler. See tokencanopy/e2a#792.
func TestTrailingSlashDoesNotDeleteAccount(t *testing.T) {
	var deletes int
	srv := testServer(t, func(d *Deps) {
		orig := d.DeleteUserData
		d.DeleteUserData = func(ctx context.Context, user *identity.User) (*identity.DeleteUserDataResult, error) {
			deletes++
			return orig(ctx, user)
		}
	})

	code, _ := sendJSON(t, http.MethodDelete, srv.URL+"/v1/account/?confirm=DELETE", "good", nil)
	if code != 404 {
		t.Fatalf("want 404 (not routed to DELETE /v1/account), got %d", code)
	}
	if deletes != 0 {
		t.Fatalf("DeleteUserData was called %d time(s), want 0", deletes)
	}

	// Sanity: the exact (non-collapsed) path still deletes, so a 404 above
	// means "not routed" and not "the whole server is broken".
	code, _ = sendJSON(t, http.MethodDelete, srv.URL+"/v1/account?confirm=DELETE", "good", nil)
	if code != 200 {
		t.Fatalf("want 200 on the exact path, got %d", code)
	}
	if deletes != 1 {
		t.Fatalf("DeleteUserData was called %d time(s), want 1", deletes)
	}
}

func TestTrailingSlashDoesNotDeleteAgent(t *testing.T) {
	var deletes int
	srv := testServer(t, func(d *Deps) {
		orig := d.DeleteAgent
		d.DeleteAgent = func(ctx context.Context, agentID, userID string) error {
			deletes++
			return orig(ctx, agentID, userID)
		}
	})

	code, _ := sendJSON(t, http.MethodDelete, srv.URL+"/v1/agents/support%40acme.com/?confirm=DELETE", "good", nil)
	if code != 404 {
		t.Fatalf("want 404 (not routed to DELETE /v1/agents/{email}), got %d", code)
	}
	if deletes != 0 {
		t.Fatalf("DeleteAgent was called %d time(s), want 0", deletes)
	}

	// Sanity: the exact (non-collapsed) path still deletes.
	code, _ = sendJSON(t, http.MethodDelete, srv.URL+"/v1/agents/support%40acme.com?confirm=DELETE", "good", nil)
	if code != 200 {
		t.Fatalf("want 200 on the exact path, got %d", code)
	}
	if deletes != 1 {
		t.Fatalf("DeleteAgent was called %d time(s), want 1", deletes)
	}
}
