package f5api

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	defaultTimeout = 10 * time.Second
	loginURL       = "/mgmt/shared/authn/login"
	logoutURL      = "/mgmt/shared/authz/tokens/"
	poolStatsURL   = "/mgmt/tm/ltm/pool/stats"
	syncStatusURL  = "/mgmt/tm/cm/sync-status"
)

// -----------------------------
// Model
// -----------------------------
type Model struct {
	User            string
	Pass            string
	Host            string
	Port            string
	MaxRetries      int
	RetryDelay      time.Duration
	InsecureSkipTLS bool
	Logger          *slog.Logger
}

// -----------------------------
// Structs for Token Handling
// -----------------------------
type F5Token struct {
	Token struct {
		Token            string `json:"token"`
		ExpirationMicros int64  `json:"expirationMicros"`
	} `json:"token"`
}

// -----------------------------
// Structs for Pool Stats
// -----------------------------
type PoolStats struct {
	Entries map[string]Entry `json:"entries"`
}

type Entry struct {
	NestedStats NestedStats `json:"nestedStats"`
}

type NestedStats struct {
	Entries StatEntries `json:"entries"`
}

type StatEntries struct {
	ActiveMemberCnt         StatValue   `json:"activeMemberCnt"`
	AvailableMemberCnt      StatValue   `json:"availableMemberCnt"`
	MemberCnt               StatValue   `json:"memberCnt"`
	MinActiveMembers        StatValue   `json:"minActiveMembers"`
	ServersideCurConns      StatValue   `json:"serverside.curConns"`
	ServersideTotConns      StatValue   `json:"serverside.totConns"`
	StatusAvailabilityState StatMessage `json:"status.availabilityState"`
	StatusEnabledState      StatMessage `json:"status.enabledState"`
	StatusStatusReason      StatMessage `json:"status.statusReason"`
	TmName                  StatMessage `json:"tmName"`
}

type StatValue struct {
	Value int64 `json:"value"`
}

type StatMessage struct {
	Description string `json:"description"`
}

// -----------------------------
// Structs for Sync Status
// -----------------------------
type SyncStatusResponse struct {
	Entries map[string]SyncEntry `json:"entries"`
}

type SyncEntry struct {
	NestedStats struct {
		Entries struct {
			Status struct {
				Description string `json:"description"`
			} `json:"status"`
		} `json:"entries"`
	} `json:"nestedStats"`
}

// -----------------------------
// Token Lifecycle
// -----------------------------

// Login creates a new token and returns its value.
func (m *Model) Login() (string, error) {
	logger := m.Logger
	host := m.Host

	payload := map[string]string{
		"username":          m.User,
		"password":          m.Pass,
		"loginProviderName": "local",
	}
	data, _ := json.Marshal(payload)

	resp, err := m.doRequest("POST", m.apiURL(loginURL), "application/json", bytes.NewReader(data), "")
	if err != nil {
		logger.Error("[f5api] login request failed",
			slog.String("host", host), slog.Any("error", err))
		return "", fmt.Errorf("login request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logger.Error("[f5api] login failed",
			slog.String("host", host), slog.Int("status", resp.StatusCode))
		return "", fmt.Errorf("login failed: HTTP %d", resp.StatusCode)
	}

	var tokenResp F5Token
	if err := decodeJSON(resp.Body, &tokenResp); err != nil {
		logger.Error("[f5api] failed to parse login response",
			slog.String("host", host), slog.Any("error", err))
		return "", fmt.Errorf("failed to parse login response: %w", err)
	}

	token := tokenResp.Token.Token
	logger.Debug("[f5api] token created",
		slog.String("host", host),
		slog.String("token", token))
	return token, nil
}

// Logout deletes a token on the F5 device.
func (m *Model) Logout(token string) {
	logger := m.Logger
	host := m.Host

	url := m.apiURL(logoutURL + token)
	authHeader := "Basic " + basicAuth(m.User, m.Pass)

	resp, err := m.doRequest("DELETE", url, "", nil, authHeader)
	if err != nil {
		logger.Warn("[f5api] logout request failed",
			slog.String("host", host), slog.String("token", token), slog.Any("error", err))
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logger.Warn("[f5api] logout returned non-200",
			slog.String("host", host), slog.String("token", token),
			slog.Int("status", resp.StatusCode))
		return
	}

	logger.Debug("[f5api] token deleted",
		slog.String("host", host),
		slog.String("token", token))
}

// -----------------------------
// API Methods
// -----------------------------

func (m *Model) GetPoolStats(token string) (PoolStats, error) {
	resp, err := m.doRequest("GET", m.apiURL(poolStatsURL), "", nil, token)
	if err != nil {
		return PoolStats{}, fmt.Errorf("pool stats request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return PoolStats{}, fmt.Errorf("failed to get pool stats: HTTP %d", resp.StatusCode)
	}

	var stats PoolStats
	if err := decodeJSON(resp.Body, &stats); err != nil {
		return PoolStats{}, fmt.Errorf("failed to decode pool stats: %w", err)
	}
	return stats, nil
}

func (m *Model) GetSyncStatus(token string) (int, error) {
	resp, err := m.doRequest("GET", m.apiURL(syncStatusURL), "", nil, token)
	if err != nil {
		return 0, fmt.Errorf("sync status request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get sync status: HTTP %d", resp.StatusCode)
	}

	var result SyncStatusResponse
	if err := decodeJSON(resp.Body, &result); err != nil {
		return 0, fmt.Errorf("failed to decode sync status: %w", err)
	}

	for _, entry := range result.Entries {
		if entry.NestedStats.Entries.Status.Description == "In Sync" {
			return 1, nil
		}
		break
	}
	return 0, nil
}

// -----------------------------
// HTTP Request Handling
// -----------------------------

func (m *Model) doRequest(method, url, contentType string, body io.Reader, authHeader string) (*http.Response, error) {
	if m.MaxRetries <= 0 {
		m.MaxRetries = 3
	}
	if m.RetryDelay <= 0 {
		m.RetryDelay = 500 * time.Millisecond
	}

	client := &http.Client{
		Timeout: defaultTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: m.InsecureSkipTLS},
		},
	}

	var lastErr error
	backoff := m.RetryDelay
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = io.ReadAll(body)
	}

	for attempt := 0; attempt <= m.MaxRetries; attempt++ {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return nil, err
		}
		req.Close = true

		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		if authHeader != "" {
			if strings.HasPrefix(authHeader, "Basic ") {
				req.Header.Set("Authorization", authHeader)
			} else {
				req.Header.Set("X-F5-Auth-Token", authHeader)
			}
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		if resp != nil {
			err := resp.Body.Close()
			if err != nil {
				return nil, err
			}
		}
		lastErr = err
		if attempt < m.MaxRetries {
			jitter := time.Duration(float64(backoff) * (0.5 + rand.Float64()*0.5))
			time.Sleep(jitter)
			backoff *= 2
		}
	}
	return nil, fmt.Errorf("request failed after %d attempts: %w", m.MaxRetries+1, lastErr)
}

// -----------------------------
// Helpers
// -----------------------------
func (m *Model) apiURL(path string) string {
	return fmt.Sprintf("https://%s:%s%s", m.Host, m.Port, path)
}

func decodeJSON(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

func basicAuth(user, pass string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
}
