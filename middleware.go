package main

import (
	"context"
	"net/http"
	"os"
	"strings"
)

type contextKey string

const (
	ctxKeyUser   contextKey = "user"
	ctxKeyGroups contextKey = "groups"
)

// authMiddleware reads Authentik forward-auth headers injected by the proxy.
// If AUTH_REQUIRED=true and X-Authentik-Username is missing, it returns 401.
// In dev mode (AUTH_REQUIRED unset or false) all requests pass through.
func authMiddleware(next http.Handler) http.Handler {
	required := os.Getenv("AUTH_REQUIRED") == "true"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.Header.Get("X-Authentik-Username")
		groups := strings.Split(r.Header.Get("X-Authentik-Groups"), ",")

		if required && username == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyUser, username)
		ctx = context.WithValue(ctx, ctxKeyGroups, groups)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requireWrite returns 403 if AUTH_REQUIRED=true and user lacks ossf-write or ossf-admin.
func requireWrite(next http.Handler) http.Handler {
	required := os.Getenv("AUTH_REQUIRED") == "true"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if required && !hasAnyGroup(r, "ossf-write", "ossf-admin") {
			http.Error(w, "Forbidden — requires ossf-write role", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireAdmin returns 403 if AUTH_REQUIRED=true and user lacks ossf-admin.
func requireAdmin(next http.Handler) http.HandlerFunc {
	required := os.Getenv("AUTH_REQUIRED") == "true"
	return func(w http.ResponseWriter, r *http.Request) {
		if required && !hasAnyGroup(r, "ossf-admin") {
			http.Error(w, "Forbidden — requires ossf-admin role", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func hasAnyGroup(r *http.Request, groups ...string) bool {
	userGroups, _ := r.Context().Value(ctxKeyGroups).([]string)
	for _, want := range groups {
		for _, g := range userGroups {
			if strings.TrimSpace(g) == want {
				return true
			}
		}
	}
	return false
}

// currentUser returns the authenticated username from context (empty in dev mode).
func currentUser(r *http.Request) string {
	u, _ := r.Context().Value(ctxKeyUser).(string)
	return u
}
