package server

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"uptimelit/internal/auth"
	"uptimelit/internal/db"
)

type authContext struct{}
type identity struct {
	User    *auth.User
	Session string
}

func current(r *http.Request) *identity {
	v, _ := r.Context().Value(authContext{}).(*identity)
	return v
}
func (s *Server) secure() bool {
	return strings.HasPrefix(strings.ToLower(s.authConfig.OIDCRedirectURL), "https://")
}
func (s *Server) setCookie(w http.ResponseWriter, r *http.Request, name, value, path string, age int) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: value, Path: path, MaxAge: age, HttpOnly: true, Secure: s.secure() || r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"), SameSite: http.SameSiteLaxMode})
}
func (s *Server) authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		raw := ""
		bearer := strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ")
		if bearer {
			raw = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		} else if c, e := r.Cookie("token"); e == nil {
			raw = c.Value
		}
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" && !bearer {
			// Browser writes must originate from this host and use the explicit JSON API header.
			if r.Header.Get("X-UptimeLit-Request") != "1" {
				respondError(w, 403, "missing CSRF protection header")
				return
			}
			if origin := r.Header.Get("Origin"); origin != "" {
				u, e := url.Parse(origin)
				if e != nil || u.Host != r.Host {
					respondError(w, 403, "invalid request origin")
					return
				}
			}
			if site := r.Header.Get("Sec-Fetch-Site"); site == "cross-site" {
				respondError(w, 403, "cross-site request denied")
				return
			}
		}
		if raw != "" {
			if u, id, e := s.auth.Authenticate(r.Context(), raw); e == nil {
				if u.AuthMethod != "password" || s.authConfig.PasswordLoginEnabled() {
					ctx := context.WithValue(r.Context(), authContext{}, &identity{u, id})
					r = r.WithContext(db.WithOwner(ctx, u.ID))
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if current(r) == nil {
			respondError(w, 401, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func requireAdmin(next http.Handler) http.Handler {
	return requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if current(r).User.Role != "admin" {
			respondError(w, 403, "administrator required")
			return
		}
		next.ServeHTTP(w, r)
	}))
}
func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if e := json.NewDecoder(r.Body).Decode(v); e != nil {
		respondError(w, 400, "invalid JSON body")
		return false
	}
	return true
}
func authError(w http.ResponseWriter, e error) {
	switch {
	case errors.Is(e, sql.ErrNoRows):
		respondError(w, 404, "user not found")
	case auth.IsConflict(e):
		respondError(w, 409, "username or identity already exists")
	default:
		respondError(w, 400, e.Error())
	}
}
func (s *Server) authRoutes(r chi.Router) {
	r.Get("/auth/config", func(w http.ResponseWriter, r *http.Request) {
		empty, e := s.auth.Empty(r.Context())
		if e != nil {
			respondError(w, 500, "could not read authentication configuration")
			return
		}
		respondJSON(w, 200, map[string]any{"passwordLoginEnabled": s.authConfig.PasswordLoginEnabled(), "oidcEnabled": s.oidc != nil, "oidcButtonTitle": s.authConfig.OIDCButtonTitle, "bootstrap": empty})
	})
	r.Post("/auth/login", s.login)
	r.Get("/auth/oidc/login", s.oidcLogin)
	r.Get("/auth/oidc/callback", s.oidcCallback)
	r.Group(func(r chi.Router) {
		r.Use(requireAuth)
		r.Get("/auth/me", func(w http.ResponseWriter, r *http.Request) { respondJSON(w, 200, current(r).User) })
		r.Post("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
			if e := s.auth.Logout(r.Context(), current(r).Session); e != nil {
				respondError(w, 500, "could not end session")
				return
			}
			s.setCookie(w, r, "token", "", "/", -1)
			respondJSON(w, 200, map[string]bool{"ok": true})
		})
		r.Put("/auth/profile", s.profile)
	})
	r.Route("/admin/users", func(r chi.Router) {
		r.Use(requireAdmin)
		r.Get("/", s.listUsers)
		r.Post("/", s.createUser)
		r.Put("/{id}", s.updateUser)
		r.Delete("/{id}", s.deleteUser)
	})
}

type attempt struct {
	Start time.Time
	Count int
}
type limiter struct {
	mu      sync.Mutex
	entries map[string]attempt
}

func (l *limiter) allow(r *http.Request) bool {
	host, _, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		host = r.RemoteAddr
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.entries == nil {
		l.entries = map[string]attempt{}
	}
	now := time.Now()
	for k, v := range l.entries {
		if now.Sub(v.Start) > time.Minute {
			delete(l.entries, k)
		}
	}
	v := l.entries[host]
	if v.Count == 0 {
		v.Start = now
	}
	v.Count++
	l.entries[host] = v
	return v.Count <= 20
}
func (s *Server) finishLogin(w http.ResponseWriter, r *http.Request, u *auth.User, method string, redirect bool) {
	token, e := s.auth.Session(r.Context(), u, method)
	if e != nil {
		respondError(w, 500, "could not create session")
		return
	}
	s.setCookie(w, r, "token", token, "/", 7*24*3600)
	u.AuthMethod = method
	if redirect {
		http.Redirect(w, r, "/", 302)
		return
	}
	respondJSON(w, 200, map[string]any{"token": token, "user": u})
}
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if !s.authConfig.PasswordLoginEnabled() {
		respondError(w, 403, "password login is disabled")
		return
	}
	if !s.limiter.allow(r) {
		w.Header().Set("Retry-After", "60")
		respondError(w, 429, "too many login attempts; retry in one minute")
		return
	}
	var req struct{ Username, Password string }
	if !decodeBody(w, r, &req) {
		return
	}
	u, e := s.auth.Login(r.Context(), req.Username, req.Password)
	if e != nil {
		respondError(w, 401, "invalid credentials or invalid first administrator password (8–72 bytes)")
		return
	}
	s.finishLogin(w, r, u, "password", false)
}
func (s *Server) profile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username        string `json:"username"`
		NewPassword     string `json:"newPassword"`
		CurrentPassword string `json:"currentPassword"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	u := current(r).User
	if u.AuthMethod == "oidc" && req.Username != u.Username {
		respondError(w, 403, "username cannot be changed during an OIDC session")
		return
	}
	if req.NewPassword != "" {
		if !s.authConfig.PasswordLoginEnabled() {
			respondError(w, 403, "password login is disabled")
			return
		}
		if u.PasswordHash != "" && !auth.PasswordMatches(u.PasswordHash, req.CurrentPassword) {
			respondError(w, 403, "current password is incorrect")
			return
		}
	}
	if e := s.auth.UpdateUser(r.Context(), u.ID, req.Username, u.Role, req.NewPassword, false, s.authConfig.PasswordLoginEnabled()); e != nil {
		authError(w, e)
		return
	}
	s.setCookie(w, r, "token", "", "/", -1)
	respondJSON(w, 200, map[string]bool{"loginRequired": true})
}
func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	u, e := s.auth.Users(r.Context())
	if e != nil {
		respondError(w, 500, "could not list users")
		return
	}
	respondJSON(w, 200, u)
}

type userRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	NewPassword string `json:"newPassword"`
	Role        string `json:"role"`
	ResetOIDC   bool   `json:"resetOIDC"`
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var q userRequest
	if !decodeBody(w, r, &q) {
		return
	}
	if q.Role == "" {
		q.Role = "user"
	}
	if q.Password == "" && s.oidc == nil {
		respondError(w, 400, "a password is required without OIDC")
		return
	}
	u, e := s.auth.CreateUser(r.Context(), q.Username, q.Password, q.Role)
	if e != nil {
		authError(w, e)
		return
	}
	respondJSON(w, 201, u)
}
func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	var q userRequest
	if !decodeBody(w, r, &q) {
		return
	}
	if e := s.auth.UpdateUser(r.Context(), chi.URLParam(r, "id"), q.Username, q.Role, q.NewPassword, q.ResetOIDC, s.authConfig.PasswordLoginEnabled()); e != nil {
		authError(w, e)
		return
	}
	respondJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	ids, e := s.auth.DeleteUser(r.Context(), chi.URLParam(r, "id"), current(r).User.ID)
	if e != nil {
		authError(w, e)
		return
	}
	for _, id := range ids {
		s.manager.Delete(id)
	}
	w.WriteHeader(204)
}
func (s *Server) oidcLogin(w http.ResponseWriter, r *http.Request) {
	if s.oidc == nil {
		respondError(w, 404, "OIDC is not configured")
		return
	}
	state, nonce, verifier := auth.Random(), auth.Random(), oauth2.GenerateVerifier()
	s.setCookie(w, r, "oidc_state", state, "/api/auth/oidc", 600)
	s.setCookie(w, r, "oidc_nonce", nonce, "/api/auth/oidc", 600)
	s.setCookie(w, r, "oidc_verifier", verifier, "/api/auth/oidc", 600)
	http.Redirect(w, r, s.oidc.AuthorizationURL(state, nonce, verifier), 302)
}
func (s *Server) oidcCallback(w http.ResponseWriter, r *http.Request) {
	if s.oidc == nil {
		respondError(w, 404, "OIDC is not configured")
		return
	}
	state, e1 := r.Cookie("oidc_state")
	nonce, e2 := r.Cookie("oidc_nonce")
	verifier, e3 := r.Cookie("oidc_verifier")
	for _, n := range []string{"oidc_state", "oidc_nonce", "oidc_verifier"} {
		s.setCookie(w, r, n, "", "/api/auth/oidc", -1)
	}
	fail := func(message string) { http.Redirect(w, r, "/?oidc_error="+url.QueryEscape(message), 302) }
	if e1 != nil || e2 != nil || e3 != nil || state.Value == "" || nonce.Value == "" || verifier.Value == "" || subtle.ConstantTimeCompare([]byte(state.Value), []byte(r.URL.Query().Get("state"))) != 1 {
		fail("OIDC session expired or is invalid")
		return
	}
	if r.URL.Query().Get("error") != "" || r.URL.Query().Get("code") == "" {
		fail("OIDC provider rejected login")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	i, e := s.oidc.Authenticate(ctx, r.URL.Query().Get("code"), nonce.Value, verifier.Value)
	if e != nil {
		fail("OIDC token validation failed")
		return
	}
	u, e := s.auth.OIDCUser(ctx, i)
	if e != nil {
		fail("Account is not provisioned or could not be linked; contact the administrator")
		return
	}
	s.finishLogin(w, r, u, "oidc", true)
}

// resourceOwner covers every monitor/notifier ID route, including history and test actions.
func (s *Server) resourceOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := chi.URLParam(r, "id")
		// chi may not have resolved URL params when group middleware runs; parse resource path.
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 3 && (parts[1] == "endpoints" || parts[1] == "notifiers") {
			if raw == "" {
				raw = parts[2]
			}
			id, e := parseResourceID(raw)
			if e != nil {
				respondError(w, 400, "invalid id")
				return
			}
			if parts[1] == "endpoints" {
				_, e = s.db.GetEndpoint(r.Context(), id)
			} else {
				_, e = s.db.GetNotifier(r.Context(), id)
			}
			if e != nil {
				respondError(w, 404, "resource not found")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func parseResourceID(v string) (int64, error) { return strconv.ParseInt(v, 10, 64) }
