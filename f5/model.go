package f5

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

const (
	defaultTimeout = 10 * time.Second
	loginURL       = "/mgmt/shared/authn/login"
	logoutURL      = "/mgmt/shared/authz/tokens/"
	poolStatsURL   = "/mgmt/tm/ltm/pool/stats"
	syncStatusURL  = "/mgmt/tm/cm/sync-status"
	syncStatusKey  = "entries.https://localhost/mgmt/tm/cm/sync-status/0.nestedStats.entries.status.description"
)

type Model struct {
	User string
	Pass string
	Host string
	Port string
}

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
	// (optional) add more fields if needed
}

type StatValue struct {
	Value int64 `json:"value"`
}

type StatMessage struct {
	Description string `json:"description"`
}

type F5Token struct {
	Token Token `json:"token"`
}

type Token struct {
	Token string `json:"token"`
}

// Authenticate logs in to the F5 device and retrieves a session token
func (m *Model) Authenticate() (string, error) {
	url := m.apiURL(loginURL)

	payload, _ := json.Marshal(map[string]string{
		"username":          m.User,
		"password":          m.Pass,
		"loginProviderName": "tmos",
	})

	resp, err := m.doRequest("POST", url, "application/json", bytes.NewBuffer(payload), "")
	if err != nil {
		return "", fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed: HTTP %d", resp.StatusCode)
	}

	var tokenResp F5Token
	if err := decodeJSON(resp.Body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse auth response: %w", err)
	}

	return tokenResp.Token.Token, nil
}

// GetPoolStats fetches pool statistics using the given session token
func (m *Model) GetPoolStats(sessionID string) (PoolStats, error) {
	url := m.apiURL(poolStatsURL)
	resp, err := m.doRequest("GET", url, "", nil, sessionID)
	if err != nil {
		return PoolStats{}, fmt.Errorf("pool stats request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PoolStats{}, fmt.Errorf("failed to get pool stats: HTTP %d", resp.StatusCode)
	}

	var stats PoolStats
	if err := decodeJSON(resp.Body, &stats); err != nil {
		return PoolStats{}, fmt.Errorf("failed to decode pool stats: %w", err)
	}

	return stats, nil
}

// GetSyncStatus returns 1 if device is in sync, 0 otherwise
func (m *Model) GetSyncStatus(sessionID string) (int, error) {
	url := m.apiURL(syncStatusURL)
	resp, err := m.doRequest("GET", url, "", nil, sessionID)
	if err != nil {
		return 0, fmt.Errorf("sync status request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get sync status: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read sync response: %w", err)
	}

	status := gjson.Get(string(body), syncStatusKey)
	if status.String() == "In Sync" {
		return 1, nil
	}
	return 0, nil
}

// Logout ends the current F5 session
func (m *Model) Logout(sessionID string) error {
	url := m.apiURL(logoutURL + sessionID)
	authHeader := "Basic " + basicAuth(m.User, m.Pass)

	resp, err := m.doRequest("DELETE", url, "", nil, authHeader)
	if err != nil {
		return fmt.Errorf("logout request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("logout failed: HTTP %d", resp.StatusCode)
	}

	return nil
}

// doRequest creates and sends an HTTP request with optional token or basic auth
func (m *Model) doRequest(method, url, contentType string, body io.Reader, authHeader string) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // you should secure this in production
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   defaultTimeout,
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Close = true
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if authHeader != "" {
		if method == "DELETE" {
			req.Header.Set("Authorization", authHeader)
		} else {
			req.Header.Set("X-F5-Auth-Token", authHeader)
		}
	}

	return client.Do(req)
}

// apiURL builds full URL from host, port, and path
func (m *Model) apiURL(path string) string {
	return fmt.Sprintf("https://%s:%s%s", m.Host, m.Port, path)
}

// decodeJSON decodes response JSON into a given struct
func decodeJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// basicAuth returns base64-encoded basic auth string
func basicAuth(user, pass string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
}
