package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	authURL  = "https://accounts.spotify.com/authorize"
	tokenURL = "https://accounts.spotify.com/api/token"
)

type OAuthConfig struct {
	ClientID    string
	RedirectURI string
	Scopes      []string
	UserAgent   string
	TLSCertFile string
	TLSKeyFile  string
}

type LoginResult struct {
	Token Token
}

func LoginPKCE(ctx context.Context, cfg OAuthConfig) (LoginResult, error) {
	if cfg.ClientID == "" {
		return LoginResult{}, fmt.Errorf("client id is required")
	}
	if cfg.RedirectURI == "" {
		return LoginResult{}, fmt.Errorf("redirect uri is required")
	}

	redirect, err := url.Parse(cfg.RedirectURI)
	if err != nil {
		return LoginResult{}, fmt.Errorf("parse redirect uri: %w", err)
	}
	if redirect.Scheme != "http" && redirect.Scheme != "https" {
		return LoginResult{}, fmt.Errorf("redirect uri scheme must be http or https (got %q)", redirect.Scheme)
	}
	if redirect.Host == "" {
		return LoginResult{}, fmt.Errorf("redirect uri must include host (got %q)", cfg.RedirectURI)
	}
	if redirect.Scheme == "https" {
		if cfg.TLSCertFile == "" || cfg.TLSKeyFile == "" {
			return LoginResult{}, fmt.Errorf("https redirect requires SPOTIFY_TLS_CERT_FILE and SPOTIFY_TLS_KEY_FILE")
		}
	}

	p, err := newPKCE()
	if err != nil {
		return LoginResult{}, err
	}

	state, err := randomURLSafe(24)
	if err != nil {
		return LoginResult{}, err
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	cbPath := redirect.EscapedPath()
	if cbPath == "" {
		// Browsers will request "/" for a bare host redirect like "http://localhost:8888".
		cbPath = "/"
	}

	mux := http.NewServeMux()
	mux.HandleFunc(cbPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if gotState := q.Get("state"); gotState != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}
		if e := q.Get("error"); e != "" {
			http.Error(w, "spotify auth error: "+e, http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("spotify auth error: %s", e):
			default:
			}
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		_, _ = io.WriteString(w, "Login complete. You can close this window.")
		select {
		case codeCh <- code:
		default:
		}
	})

	listener, server, err := startCallbackServer(mux, redirect.Scheme, redirect.Host, cfg.TLSCertFile, cfg.TLSKeyFile)
	if err != nil {
		return LoginResult{}, err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		_ = listener.Close()
	}()

	redirectURIForSpotify := cfg.RedirectURI
	if isPortZero(redirect.Host) {
		// Only rewrite the port if the caller explicitly used ":0".
		// Keep the original hostname (e.g. "localhost") so it can still match a registered redirect.
		port := listener.Addr().(*net.TCPAddr).Port
		r := *redirect
		r.Host = net.JoinHostPort(redirect.Hostname(), strconv.Itoa(port))
		redirectURIForSpotify = r.String()
	}

	authorize := buildAuthorizeURL(cfg, redirectURIForSpotify, state, p.Challenge)
	if err := openBrowser(authorize); err != nil {
		// Not fatal; print URL is caller's responsibility.
		return LoginResult{}, fmt.Errorf("open browser: %w", err)
	}

	select {
	case <-ctx.Done():
		return LoginResult{}, ctx.Err()
	case err := <-errCh:
		return LoginResult{}, err
	case code := <-codeCh:
		tok, err := exchangeCode(ctx, cfg, redirectURIForSpotify, code, p.Verifier)
		if err != nil {
			return LoginResult{}, err
		}
		return LoginResult{Token: tok}, nil
	}
}

func isPortZero(hostport string) bool {
	_, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return false
	}
	return port == "0"
}

func startCallbackServer(handler http.Handler, scheme, host, certFile, keyFile string) (net.Listener, *http.Server, error) {
	if host == "" {
		host = "localhost:8888"
	}

	ln, err := net.Listen("tcp", host)
	if err != nil {
		return nil, nil, fmt.Errorf("listen %s: %w", host, err)
	}

	server := &http.Server{Handler: handler}
	go func() {
		if scheme == "https" {
			_ = server.ServeTLS(ln, certFile, keyFile)
			return
		}
		_ = server.Serve(ln)
	}()

	return ln, server, nil
}

func buildAuthorizeURL(cfg OAuthConfig, redirectURI, state, codeChallenge string) string {
	q := url.Values{}
	q.Set("client_id", cfg.ClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("code_challenge_method", "S256")
	q.Set("code_challenge", codeChallenge)
	if len(cfg.Scopes) > 0 {
		q.Set("scope", strings.Join(cfg.Scopes, " "))
	}

	u, _ := url.Parse(authURL)
	u.RawQuery = q.Encode()
	return u.String()
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type tokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func exchangeCode(ctx context.Context, cfg OAuthConfig, redirectURI, code, codeVerifier string) (Token, error) {
	form := url.Values{}
	form.Set("client_id", cfg.ClientID)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cfg.UserAgent != "" {
		req.Header.Set("User-Agent", cfg.UserAgent)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var te tokenError
		_ = json.Unmarshal(body, &te)
		if te.Error != "" {
			return Token{}, fmt.Errorf("token exchange failed: %s (%s)", te.Error, te.ErrorDescription)
		}
		return Token{}, fmt.Errorf("token exchange failed: http %d", resp.StatusCode)
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return Token{}, fmt.Errorf("decode token response: %w", err)
	}
	if tr.AccessToken == "" {
		return Token{}, errors.New("missing access_token in response")
	}

	return Token{
		AccessToken:  tr.AccessToken,
		TokenType:    tr.TokenType,
		Scope:        tr.Scope,
		RefreshToken: tr.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second),
	}, nil
}

func Refresh(ctx context.Context, cfg OAuthConfig, refreshToken string) (Token, error) {
	if refreshToken == "" {
		return Token{}, errors.New("missing refresh token")
	}

	form := url.Values{}
	form.Set("client_id", cfg.ClientID)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cfg.UserAgent != "" {
		req.Header.Set("User-Agent", cfg.UserAgent)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var te tokenError
		_ = json.Unmarshal(body, &te)
		if te.Error != "" {
			return Token{}, fmt.Errorf("token refresh failed: %s (%s)", te.Error, te.ErrorDescription)
		}
		return Token{}, fmt.Errorf("token refresh failed: http %d", resp.StatusCode)
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return Token{}, fmt.Errorf("decode refresh response: %w", err)
	}
	if tr.AccessToken == "" {
		return Token{}, errors.New("missing access_token in refresh response")
	}

	newRefresh := refreshToken
	if tr.RefreshToken != "" {
		newRefresh = tr.RefreshToken
	}

	return Token{
		AccessToken:  tr.AccessToken,
		TokenType:    tr.TokenType,
		Scope:        tr.Scope,
		RefreshToken: newRefresh,
		ExpiresAt:    time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second),
	}, nil
}

func openBrowser(u string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	return cmd.Start()
}
