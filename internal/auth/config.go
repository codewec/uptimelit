package auth

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	JWTSecret                                                                    string
	OIDCIssuer, OIDCClientID, OIDCClientSecret, OIDCRedirectURL, OIDCButtonTitle string
	OIDCInsecureSkipTLSVerify, DisablePasswordLogin                              bool
}

func (c Config) OIDCEnabled() bool          { return c.OIDCIssuer != "" && c.OIDCClientID != "" }
func (c Config) PasswordLoginEnabled() bool { return !c.DisablePasswordLogin || !c.OIDCEnabled() }
func LoadConfig() (Config, error) {
	c := Config{JWTSecret: os.Getenv("JWT_SECRET"), OIDCIssuer: strings.TrimSpace(os.Getenv("OIDC_ISSUER")), OIDCClientID: strings.TrimSpace(os.Getenv("OIDC_CLIENT_ID")), OIDCClientSecret: os.Getenv("OIDC_CLIENT_SECRET"), OIDCRedirectURL: os.Getenv("OIDC_REDIRECT_URL"), OIDCButtonTitle: os.Getenv("OIDC_BUTTON_TITLE"), OIDCInsecureSkipTLSVerify: os.Getenv("OIDC_INSECURE_SKIP_TLS_VERIFY") == "true", DisablePasswordLogin: os.Getenv("DISABLE_PASSWORD_LOGIN") == "true"}
	if c.OIDCButtonTitle == "" {
		c.OIDCButtonTitle = "Sign in with OIDC"
	}
	if (c.OIDCIssuer == "") != (c.OIDCClientID == "") {
		return c, fmt.Errorf("OIDC_ISSUER and OIDC_CLIENT_ID must both be set")
	}
	if c.OIDCEnabled() {
		u, e := url.Parse(c.OIDCRedirectURL)
		if e != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return c, fmt.Errorf("OIDC_REDIRECT_URL must be an absolute HTTP(S) callback URL")
		}
	}
	if c.JWTSecret != "" && len(c.JWTSecret) < 32 {
		return c, fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}
	return c, nil
}
