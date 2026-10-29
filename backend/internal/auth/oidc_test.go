package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	jose "github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

func TestOIDCProtocol(t *testing.T) {
	key, e := rsa.GenerateKey(rand.Reader, 2048)
	if e != nil {
		t.Fatal(e)
	}
	var mu sync.Mutex
	mode := "valid"
	var issuer string
	pkce := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(map[string]any{"issuer": issuer, "authorization_endpoint": issuer + "/authorize", "token_endpoint": issuer + "/token", "jwks_uri": issuer + "/keys", "id_token_signing_alg_values_supported": []string{"RS256"}})
		case "/keys":
			json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "key", Algorithm: "RS256", Use: "sig"}}})
		case "/token":
			r.ParseForm()
			mu.Lock()
			m := mode
			pkce = r.Form.Get("code_verifier")
			mu.Unlock()
			claims := jwt.MapClaims{"iss": issuer, "sub": "user-subject", "aud": "client", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(), "nonce": "expected-nonce", "preferred_username": "alice"}
			switch m {
			case "nonce":
				claims["nonce"] = "wrong"
			case "audience":
				claims["aud"] = "other-client"
			case "issuer":
				claims["iss"] = "https://other-issuer"
			case "expired":
				claims["exp"] = time.Now().Add(-time.Hour).Unix()
			case "subject":
				claims["sub"] = ""
			}
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
			token.Header["kid"] = "key"
			signKey := key
			if m == "signature" {
				signKey, _ = rsa.GenerateKey(rand.Reader, 2048)
			}
			raw, err := token.SignedString(signKey)
			if err != nil {
				t.Error(err)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"access_token": "access", "token_type": "Bearer", "id_token": raw})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	issuer = srv.URL
	cfg := Config{OIDCIssuer: issuer, OIDCClientID: "client", OIDCClientSecret: "secret", OIDCRedirectURL: "https://app.example/api/auth/oidc/callback"}
	a, e := NewOIDCAuthenticator(context.Background(), &cfg)
	if e != nil {
		t.Fatal(e)
	}
	verifier := oauth2.GenerateVerifier()
	u, _ := url.Parse(a.AuthorizationURL("state", "expected-nonce", verifier))
	q := u.Query()
	if q.Get("state") != "state" || q.Get("nonce") != "expected-nonce" || q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") != oauth2.S256ChallengeFromVerifier(verifier) {
		t.Fatal("missing OIDC flow protections", q)
	}
	for _, m := range []string{"valid", "nonce", "audience", "issuer", "expired", "signature", "subject"} {
		t.Run(m, func(t *testing.T) {
			mu.Lock()
			mode = m
			mu.Unlock()
			i, e := a.Authenticate(context.Background(), "code", "expected-nonce", verifier)
			if m == "valid" {
				if e != nil || i.Subject != "user-subject" || i.PreferredUsername != "alice" {
					t.Fatalf("identity=%+v error=%v", i, e)
				}
			} else if e == nil {
				t.Fatal("invalid token accepted")
			}
		})
	}
	mu.Lock()
	defer mu.Unlock()
	if pkce != verifier {
		t.Fatal("PKCE verifier not sent")
	}
}

func TestOIDCTLSVerification(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		issuer := "https://" + r.Host
		json.NewEncoder(w).Encode(map[string]any{"issuer": issuer, "authorization_endpoint": issuer + "/authorize", "token_endpoint": issuer + "/token", "jwks_uri": issuer + "/keys", "id_token_signing_alg_values_supported": []string{"RS256"}})
	}))
	defer srv.Close()
	cfg := Config{OIDCIssuer: srv.URL, OIDCClientID: "client", OIDCRedirectURL: "https://app.example/api/auth/oidc/callback"}
	if _, e := NewOIDCAuthenticator(context.Background(), &cfg); e == nil {
		t.Fatal("untrusted certificate accepted by default")
	}
	cfg.OIDCInsecureSkipTLSVerify = true
	if _, e := NewOIDCAuthenticator(context.Background(), &cfg); e != nil {
		t.Fatal("explicit TLS opt-out failed", e)
	}
}
