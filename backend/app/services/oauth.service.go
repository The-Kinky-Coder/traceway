package services

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tracewayapp/traceway/backend/app/config"
	traceway "go.tracewayapp.com"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/openidConnect"
)

type oauthService struct {
	googleEnabled   bool
	githubEnabled   bool
	oidcEnabled     bool
	oidcDisplayName string
	oidcAutoCreate  bool
	oidcOrgClaim    string
}

var OAuthService *oauthService

func InitOAuth() {
	cfg := config.Config

	svc := &oauthService{
		googleEnabled:   cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "",
		githubEnabled:   cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "",
		oidcEnabled:     cfg.OIDCClientID != "" && cfg.OIDCClientSecret != "" && cfg.OIDCDiscoveryURL != "",
		oidcDisplayName: cfg.OIDCDisplayName,
		oidcAutoCreate:  cfg.OIDCAutoCreateUsers == "true",
		oidcOrgClaim:    cfg.OIDCOrgClaim,
	}
	OAuthService = svc

	if !svc.googleEnabled && !svc.githubEnabled && !svc.oidcEnabled {
		return
	}

	secret := cfg.OAuthSessionSecret
	if secret == "" {
		secret = cfg.JWTSecret
	}

	// FilesystemStore keeps session data in /tmp (only a session ID in the cookie).
	// MaxLength(0) removes the securecookie size cap on the file contents — OIDC ID
	// tokens from providers like Authentik exceed the default 4096-byte limit.
	store := sessions.NewFilesystemStore("", []byte(secret))
	store.MaxLength(0)
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		MaxAge:   600,
		Secure:   strings.HasPrefix(cfg.AppBaseURL, "https://"),
		SameSite: http.SameSiteLaxMode,
	}
	gothic.Store = store

	providers := []goth.Provider{}
	base := strings.TrimRight(cfg.AppBaseURL, "/")
	if svc.googleEnabled {
		providers = append(providers, google.New(
			cfg.GoogleClientID,
			cfg.GoogleClientSecret,
			base+"/api/auth/callback/google",
			"email", "profile",
		))
	}
	if svc.githubEnabled {
		providers = append(providers, github.New(
			cfg.GitHubClientID,
			cfg.GitHubClientSecret,
			base+"/api/auth/callback/github",
			"user:email",
		))
	}
	if svc.oidcEnabled {
		scopes := []string{"openid", "email", "profile"}
		for _, s := range strings.Split(cfg.OIDCExtraScopes, ",") {
			if s = strings.TrimSpace(s); s != "" {
				scopes = append(scopes, s)
			}
		}
		oidcProvider, err := openidConnect.New(
			cfg.OIDCClientID,
			cfg.OIDCClientSecret,
			base+"/api/auth/callback/oidc",
			cfg.OIDCDiscoveryURL,
			scopes...,
		)
		if err != nil {
			traceway.CaptureException(fmt.Errorf("OIDC provider init failed (discovery URL may be unreachable): %w", err))
			svc.oidcEnabled = false
		} else {
			providers = append(providers, oidcProvider)
		}
	}
	goth.UseProviders(providers...)
}

func (s *oauthService) IsEnabled() bool {
	return s.googleEnabled || s.githubEnabled || s.oidcEnabled
}

func (s *oauthService) IsProviderEnabled(name string) bool {
	switch name {
	case "google":
		return s.googleEnabled
	case "github":
		return s.githubEnabled
	case "oidc":
		return s.oidcEnabled
	}
	return false
}

func (s *oauthService) EnabledProviders() []string {
	out := []string{}
	if s.googleEnabled {
		out = append(out, "google")
	}
	if s.githubEnabled {
		out = append(out, "github")
	}
	if s.oidcEnabled {
		out = append(out, "oidc")
	}
	return out
}

func (s *oauthService) OIDCAutoCreateEnabled() bool {
	return s.oidcAutoCreate
}

func (s *oauthService) OIDCDisplayName() string {
	return s.oidcDisplayName
}

func (s *oauthService) OIDCOrgClaim() string {
	return s.oidcOrgClaim
}
