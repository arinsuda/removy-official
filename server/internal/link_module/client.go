package link_module

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/arinsuda/movie-hub/config"
)

var ErrRemoteRejected = errors.New("remote rejected request")

// remoteClient makes signed server-to-server calls to the other platform.
type remoteClient struct {
	cfg  config.LinkConfig
	http *http.Client
}

func newRemoteClient(cfg config.LinkConfig) *remoteClient {
	return &remoteClient{cfg: cfg, http: &http.Client{Timeout: 5 * time.Second}}
}

func remoteAPIPrefix(platform string) string {
	if platform == PlatformPlanzi {
		return "/api/v1"
	}
	return "/api"
}

func (c *remoteClient) post(ctx context.Context, remotePlatform, path string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	url := c.cfg.RemoteAPIURL + remoteAPIPrefix(remotePlatform) + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(HeaderTimestamp, ts)
	req.Header.Set(HeaderSignature, Sign(c.cfg.Secrets[0], ts, body))
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("link remote call failed: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", ErrRemoteRejected, resp.StatusCode)
	}
	return json.Unmarshal(raw, out)
}

func (c *remoteClient) Introspect(ctx context.Context, remotePlatform, token string) (*RemoteUser, error) {
	var u RemoteUser
	if err := c.post(ctx, remotePlatform, "/internal/link/introspect", map[string]string{"token": token}, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (c *remoteClient) VerifyPassword(ctx context.Context, remotePlatform, email, password string) (*RemoteUser, error) {
	var out struct {
		OK   bool        `json:"ok"`
		User *RemoteUser `json:"user"`
	}
	if err := c.post(ctx, remotePlatform, "/internal/link/verify-password", map[string]string{"email": email, "password": password}, &out); err != nil {
		return nil, err
	}
	if !out.OK || out.User == nil {
		return nil, nil
	}
	return out.User, nil
}
