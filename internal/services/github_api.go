package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GitHubAPI handles direct communication with GitHub API
type GitHubAPI struct {
	client    *http.Client
	baseURL   string
	userAgent string
}

// GitHubAPIRelease represents a GitHub release from API
type GitHubAPIRelease struct {
	ID          int64     `json:"id"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		ID                 int64  `json:"id"`
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
		ContentType        string `json:"content_type"`
	} `json:"assets"`
}

// GitHubAPIRepository represents a GitHub repository from API
type GitHubAPIRepository struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     *string   `json:"description"`
	Private         bool      `json:"private"`
	Fork            bool      `json:"fork"`
	DefaultBranch   string    `json:"default_branch"`
	Language        *string   `json:"language"`
	StargazersCount int       `json:"stargazers_count"`
	ForksCount      int       `json:"forks_count"`
	OpenIssuesCount int       `json:"open_issues_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PushedAt        time.Time `json:"pushed_at"`
}

// GitHubAPIWebhook represents webhook configuration
type GitHubAPIWebhook struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Active bool     `json:"active"`
	Events []string `json:"events"`
	Config struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
		Secret      string `json:"secret"`
		InsecureSSL string `json:"insecure_ssl"`
	} `json:"config"`
}

// NewGitHubAPI creates a new GitHub API client
func NewGitHubAPI() *GitHubAPI {
	return &GitHubAPI{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://api.github.com",
		userAgent: "VerTree/1.0",
	}
}

// GetReleases fetches releases from a GitHub repository
func (api *GitHubAPI) GetReleases(owner, repo, token string, perPage, page int) ([]GitHubAPIRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=%d&page=%d",
		api.baseURL, owner, repo, perPage, page)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", api.userAgent)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var releases []GitHubAPIRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return releases, nil
}

// GetLatestRelease fetches the latest release from a GitHub repository
func (api *GitHubAPI) GetLatestRelease(owner, repo, token string) (*GitHubAPIRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", api.baseURL, owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", api.userAgent)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("no releases found")
		}
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var release GitHubAPIRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &release, nil
}

// CreateWebhook creates a webhook for a GitHub repository
func (api *GitHubAPI) CreateWebhook(owner, repo, token, webhookURL, secret string) (*GitHubAPIWebhook, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/hooks", api.baseURL, owner, repo)

	webhookConfig := map[string]interface{}{
		"name":   "web",
		"active": true,
		"events": []string{"release"},
		"config": map[string]string{
			"url":          webhookURL,
			"content_type": "json",
			"secret":       secret,
			"insecure_ssl": "0",
		},
	}

	jsonData, err := json.Marshal(webhookConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal webhook config: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", api.userAgent)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var webhook GitHubAPIWebhook
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &webhook, nil
}

// UpdateWebhook updates a webhook for a GitHub repository
func (api *GitHubAPI) UpdateWebhook(owner, repo, token string, webhookID int64, webhookURL, secret string) (*GitHubAPIWebhook, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/hooks/%d", api.baseURL, owner, repo, webhookID)

	webhookConfig := map[string]interface{}{
		"active": true,
		"events": []string{"release"},
		"config": map[string]string{
			"url":          webhookURL,
			"content_type": "json",
			"secret":       secret,
			"insecure_ssl": "0",
		},
	}

	jsonData, err := json.Marshal(webhookConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal webhook config: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", api.userAgent)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var webhook GitHubAPIWebhook
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &webhook, nil
}

// DeleteWebhook deletes a webhook from a GitHub repository
func (api *GitHubAPI) DeleteWebhook(owner, repo, token string, webhookID int64) error {
	url := fmt.Sprintf("%s/repos/%s/%s/hooks/%d", api.baseURL, owner, repo, webhookID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", api.userAgent)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err := api.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetRepository retrieves repository information from GitHub API
func (api *GitHubAPI) GetRepository(owner, repo, token string) (*GitHubAPIRepository, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", api.baseURL, owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", api.userAgent)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("repository not found or access denied")
		}
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("invalid token or insufficient permissions")
		}
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var repository GitHubAPIRepository
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(body, &repository); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &repository, nil
}

// TestConnection tests the connection to GitHub API with the provided token
func (api *GitHubAPI) TestConnection(owner, repo, token string) error {
	_, err := api.GetRepository(owner, repo, token)
	return err
}

// GetRateLimit gets the current rate limit status
func (api *GitHubAPI) GetRateLimit(token string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/rate_limit", api.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", api.userAgent)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// ConvertToWebhookPayload converts GitHub API release to webhook payload format
func ConvertToWebhookPayload(apiRelease *GitHubAPIRelease, owner, repo string) *WebhookPayload {
	payload := &WebhookPayload{
		Action: "published",
	}

	payload.Release.ID = apiRelease.ID
	payload.Release.TagName = apiRelease.TagName
	payload.Release.Name = apiRelease.Name
	payload.Release.Body = apiRelease.Body
	payload.Release.Prerelease = apiRelease.Prerelease
	payload.Release.Draft = apiRelease.Draft
	payload.Release.PublishedAt = apiRelease.PublishedAt

	// Convert assets
	for _, asset := range apiRelease.Assets {
		payload.Release.Assets = append(payload.Release.Assets, struct {
			ID                 int64  `json:"id"`
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		}{
			ID:                 asset.ID,
			Name:               asset.Name,
			BrowserDownloadURL: asset.BrowserDownloadURL,
			Size:               asset.Size,
		})
	}

	payload.Repository.Name = repo
	payload.Repository.FullName = fmt.Sprintf("%s/%s", owner, repo)
	payload.Repository.Owner.Login = owner

	return payload
}

// WebhookPayload represents the webhook payload structure from GitHub
type WebhookPayload struct {
	Action  string `json:"action"`
	Release struct {
		ID          int64     `json:"id"`
		TagName     string    `json:"tag_name"`
		Name        string    `json:"name"`
		Body        string    `json:"body"`
		Prerelease  bool      `json:"prerelease"`
		Draft       bool      `json:"draft"`
		PublishedAt time.Time `json:"published_at"`
		Assets      []struct {
			ID                 int64  `json:"id"`
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	} `json:"release"`
	Repository struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
		Name     string `json:"name"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}
