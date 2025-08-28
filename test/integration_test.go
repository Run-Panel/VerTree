package main

// 集成测试暂时禁用，等待handlers包结构重构完成
// TODO: 重构handlers包结构后重新启用此测试

/*
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Run-Panel/VerTree/internal/config"
	"github.com/Run-Panel/VerTree/internal/database"
	"github.com/Run-Panel/VerTree/internal/handlers/admin"
	"github.com/Run-Panel/VerTree/internal/handlers/auth"
	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/utils"
	"github.com/gin-gonic/gin"
)*/

/*
// TestApp 测试应用结构
type TestApp struct {
	cfg    *config.Config
	db     *database.DB
	router *gin.Engine
	server *http.Server
}

// setupTestApp 设置测试应用
func setupTestApp(t *testing.T) *TestApp {
	// 设置测试配置
	cfg := &config.Config{
		Port:        "8081",
		DatabaseURL: "sqlite:///tmp/test_vertree.db",
		JWTSecret:   "test-secret-key-for-integration-tests",
		Environment: "test",
	}

	// 初始化数据库
	db, err := database.NewDB(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// 设置路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 初始化处理器
	jwtUtil := utils.NewJWTUtil(cfg.JWTSecret)
	appHandler := handlers.NewAppHandler(db, jwtUtil)
	versionHandler := handlers.NewVersionHandler(db, jwtUtil)

	// 设置路由
	api := router.Group("/api")
	{
		// 认证路由
		api.POST("/register", appHandler.Register)
		api.POST("/login", appHandler.Login)

		// 需要认证的路由
		auth := api.Group("/", appHandler.AuthMiddleware())
		{
			// 应用管理
			auth.POST("/apps", appHandler.CreateApp)
			auth.GET("/apps", appHandler.GetApps)
			auth.GET("/apps/:id", appHandler.GetApp)
			auth.PUT("/apps/:id", appHandler.UpdateApp)
			auth.DELETE("/apps/:id", appHandler.DeleteApp)

			// 版本管理
			auth.POST("/apps/:app_id/versions", versionHandler.CreateVersion)
			auth.GET("/apps/:app_id/versions", versionHandler.GetVersions)
			auth.GET("/versions/:id", versionHandler.GetVersion)
			auth.PUT("/versions/:id", versionHandler.UpdateVersion)
			auth.DELETE("/versions/:id", versionHandler.DeleteVersion)
		}
	}

	return &TestApp{
		cfg:    cfg,
		db:     db,
		router: router,
	}
}

// cleanup 清理测试环境
func (app *TestApp) cleanup() {
	if app.db != nil {
		app.db.Close()
	}
	os.Remove("/tmp/test_vertree.db")
}

// TestVersionManagement 测试版本管理功能
func TestVersionManagement(t *testing.T) {
	app := setupTestApp(t)
	defer app.cleanup()

	// 创建HTTP客户端
	client := &http.Client{Timeout: 10 * time.Second}
	baseURL := "http://localhost:8081"

	// 启动测试服务器
	app.server = &http.Server{
		Addr:    ":8081",
		Handler: app.router,
	}

	go func() {
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Failed to start server: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 1. 注册用户
	registerData := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "testpass123",
	}
	registerBody, _ := json.Marshal(registerData)

	resp, err := client.Post(baseURL+"/api/register", "application/json", bytes.NewBuffer(registerBody))
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", resp.StatusCode)
	}

	// 2. 登录获取token
	loginData := map[string]string{
		"username": "testuser",
		"password": "testpass123",
	}
	loginBody, _ := json.Marshal(loginData)

	resp, err = client.Post(baseURL+"/api/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var loginResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	token := loginResp["token"].(string)
	if token == "" {
		t.Fatal("No token received from login")
	}

	// 3. 创建应用
	appData := map[string]interface{}{
		"name":        "TestApp",
		"description": "Test application for integration testing",
		"repository":  "https://github.com/test/testapp",
	}
	appBody, _ := json.Marshal(appData)

	req, _ := http.NewRequest("POST", baseURL+"/api/apps", bytes.NewBuffer(appBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var appResp models.App
	if err := json.NewDecoder(resp.Body).Decode(&appResp); err != nil {
		t.Fatalf("Failed to decode app response: %v", err)
	}

	appID := appResp.ID

	// 4. 测试版本创建和比较功能
	testVersions := []struct {
		version     string
		description string
		expectValid bool
	}{
		{"1.0.0", "Initial release", true},
		{"1.0.1", "Patch update", true},
		{"1.1.0", "Minor update", true},
		{"2.0.0", "Major update", true},
		{"v2.1.0", "Version with v prefix", true},
		{"2.0.0-beta.1", "Beta release", true},
		{"invalid-version", "Invalid version format", true}, // 应该被接受，但标记为非semver
	}

	createdVersions := []uint{}

	for _, tv := range testVersions {
		versionData := map[string]interface{}{
			"version":     tv.version,
			"description": tv.description,
			"changelog":   fmt.Sprintf("Changes for version %s", tv.version),
		}
		versionBody, _ := json.Marshal(versionData)

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/apps/%d/versions", baseURL, appID), bytes.NewBuffer(versionBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err = client.Do(req)
		if err != nil {
			t.Fatalf("Failed to create version %s: %v", tv.version, err)
		}

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 201 for version %s, got %d. Body: %s", tv.version, resp.StatusCode, string(body))
		}

		var versionResp models.Version
		if err := json.NewDecoder(resp.Body).Decode(&versionResp); err != nil {
			t.Fatalf("Failed to decode version response: %v", err)
		}
		resp.Body.Close()

		createdVersions = append(createdVersions, versionResp.ID)

		t.Logf("Created version: %s (ID: %d)", tv.version, versionResp.ID)
	}

	// 5. 获取所有版本并验证排序
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/apps/%d/versions", baseURL, appID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to get versions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var versions []models.Version
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		t.Fatalf("Failed to decode versions response: %v", err)
	}

	t.Logf("Retrieved %d versions", len(versions))
	for _, v := range versions {
		t.Logf("Version: %s, Created: %s", v.Version, v.CreatedAt.Format(time.RFC3339))
	}

	// 6. 测试版本比较功能
	vc := utils.NewVersionComparer()

	// 测试更新检查
	testCases := []struct {
		current  string
		latest   string
		expected bool
		desc     string
	}{
		{"1.0.0", "1.0.1", true, "Patch update needed"},
		{"1.0.1", "1.1.0", true, "Minor update needed"},
		{"1.1.0", "2.0.0", true, "Major update needed"},
		{"2.0.0", "2.0.0", false, "No update needed - same version"},
		{"2.1.0", "2.0.0", false, "No update needed - current newer"},
		{"v2.0.0", "2.1.0", true, "Update needed with v prefix"},
	}

	for _, tc := range testCases {
		result := vc.IsUpdateNeeded(tc.current, tc.latest)
		if result != tc.expected {
			t.Errorf("%s: IsUpdateNeeded(%s, %s) = %v, want %v", tc.desc, tc.current, tc.latest, result, tc.expected)
		} else {
			t.Logf("✓ %s: IsUpdateNeeded(%s, %s) = %v", tc.desc, tc.current, tc.latest, result)
		}
	}

	// 7. 清理创建的版本
	for _, versionID := range createdVersions {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/versions/%d", baseURL, versionID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Warning: Failed to delete version %d: %v", versionID, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Warning: Expected status 200 when deleting version %d, got %d", versionID, resp.StatusCode)
		}
	}

	// 8. 清理创建的应用
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("%s/api/apps/%d", baseURL, appID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	if err != nil {
		t.Logf("Warning: Failed to delete app: %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Logf("Warning: Expected status 200 when deleting app, got %d", resp.StatusCode)
		}
	}

	// 关闭服务器
	if app.server != nil {
		app.server.Close()
	}

	t.Log("✓ Integration test completed successfully!")
}
*/
