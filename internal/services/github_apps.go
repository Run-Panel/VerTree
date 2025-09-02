package services

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GitHubAppsAuth handles GitHub Apps authentication
type GitHubAppsAuth struct {
	AppID      int64
	PrivateKey *rsa.PrivateKey
	client     *http.Client
}

// GitHubAppInstallation represents a GitHub App installation
type GitHubAppInstallation struct {
	ID          int64             `json:"id"`
	Account     Account           `json:"account"`
	Permissions map[string]string `json:"permissions"`
	Events      []string          `json:"events"`
}

// Account represents the account that installed the GitHub App
type Account struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Type  string `json:"type"` // "User" or "Organization"
}

// InstallationToken represents a GitHub Apps installation token
type InstallationToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewGitHubAppsAuth creates a new GitHub Apps authentication client
func NewGitHubAppsAuth(appID int64, privateKeyPEM string) (*GitHubAppsAuth, error) {
	// Parse private key
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format if PKCS1 fails
		pkcs8Key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse private key (PKCS1: %v, PKCS8: %v)", err, err2)
		}
		var ok bool
		privateKey, ok = pkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not an RSA key")
		}
	}

	return &GitHubAppsAuth{
		AppID:      appID,
		PrivateKey: privateKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GenerateJWT generates a JWT token for GitHub Apps authentication
func (auth *GitHubAppsAuth) GenerateJWT() (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"iss": auth.AppID,
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(), // JWT tokens expire after 10 minutes
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(auth.PrivateKey)
}

// GetInstallations retrieves all installations for the GitHub App
func (auth *GitHubAppsAuth) GetInstallations() ([]GitHubAppInstallation, error) {
	jwtToken, err := auth.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	req, err := http.NewRequest("GET", "https://api.github.com/app/installations", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	req.Header.Set("User-Agent", "VerTree/1.0")

	resp, err := auth.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var installations []GitHubAppInstallation
	if err := json.NewDecoder(resp.Body).Decode(&installations); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return installations, nil
}

// GetInstallationToken generates an installation access token
func (auth *GitHubAppsAuth) GetInstallationToken(installationID int64) (*InstallationToken, error) {
	jwtToken, err := auth.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	req.Header.Set("User-Agent", "VerTree/1.0")

	resp, err := auth.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var token InstallationToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &token, nil
}

// GetInstallationForRepository finds the installation that has access to a specific repository
func (auth *GitHubAppsAuth) GetInstallationForRepository(owner, repo string) (*GitHubAppInstallation, error) {
	installations, err := auth.GetInstallations()
	if err != nil {
		return nil, err
	}

	// Check each installation to see if it has access to the repository
	for _, installation := range installations {
		if installation.Account.Login == owner {
			return &installation, nil
		}
	}

	return nil, fmt.Errorf("no installation found for repository %s/%s", owner, repo)
}

// GetTokenForRepository gets an installation token for accessing a specific repository
func (auth *GitHubAppsAuth) GetTokenForRepository(owner, repo string) (string, error) {
	installation, err := auth.GetInstallationForRepository(owner, repo)
	if err != nil {
		return "", err
	}

	token, err := auth.GetInstallationToken(installation.ID)
	if err != nil {
		return "", err
	}

	return token.Token, nil
}

// TestConnection tests if the GitHub App can access a repository
func (auth *GitHubAppsAuth) TestConnection(owner, repo string) error {
	token, err := auth.GetTokenForRepository(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to get installation token: %w", err)
	}

	// Test the connection using the installation token
	githubAPI := NewGitHubAPI()
	return githubAPI.TestConnection(owner, repo, token)
}
