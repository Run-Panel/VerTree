package admin

import (
	"net/http"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/gin-gonic/gin"
)

// APIDocsHandler handles API documentation endpoints
type APIDocsHandler struct{}

// NewAPIDocsHandler creates a new API docs handler
func NewAPIDocsHandler() *APIDocsHandler {
	return &APIDocsHandler{}
}

// GetAPIDocs handles GET /admin/api/v1/docs
func (h *APIDocsHandler) GetAPIDocs(c *gin.Context) {
	docs := map[string]interface{}{
		"title":       "VerTree API Documentation",
		"version":     "1.0.0",
		"description": "API documentation for VerTree version management system",
		"base_url":    "/api/v1",
		"authentication": map[string]interface{}{
			"type":        "Bearer Token",
			"format":      "Bearer <app_id>:<api_key>",
			"description": "所有客户端API都需要使用应用ID和API密钥进行认证",
			"example":     "Bearer app_abc123:sk_def456789",
		},
		"endpoints": []map[string]interface{}{
			{
				"name":        "检查更新",
				"method":      "POST",
				"path":        "/check-update",
				"permission":  "check_update",
				"description": "检查应用是否有可用的更新版本",
				"request": map[string]interface{}{
					"headers": map[string]string{
						"Authorization": "Bearer <app_id>:<api_key>",
						"Content-Type":  "application/json",
					},
					"body": map[string]interface{}{
						"app_id":          "string (required) - 应用ID，与Authorization中的app_id一致",
						"current_version": "string (required) - 当前版本号，如 'v1.2.0'",
						"channel":         "string (required) - 更新通道: stable, beta, alpha",
						"client_id":       "string (required) - 客户端唯一标识",
						"region":          "string (optional) - 客户端地区，如 'cn', 'us'",
						"arch":            "string (optional) - 系统架构，如 'amd64', 'arm64'",
						"os":              "string (optional) - 操作系统，如 'linux', 'windows', 'darwin'",
					},
					"example": map[string]interface{}{
						"app_id":          "app_abc123def456",
						"current_version": "v1.2.0",
						"channel":         "stable",
						"client_id":       "client_unique_id_12345",
						"region":          "cn",
						"arch":            "amd64",
						"os":              "linux",
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "检查成功",
						"example_no_update": map[string]interface{}{
							"code":    200,
							"message": "success",
							"data": map[string]interface{}{
								"has_update": false,
							},
						},
						"example_has_update": map[string]interface{}{
							"code":    200,
							"message": "success",
							"data": map[string]interface{}{
								"has_update":          true,
								"latest_version":      "v1.2.3",
								"download_url":        "https://releases.example.com/v1.2.3/app-linux-amd64",
								"file_size":           25165824,
								"file_checksum":       "sha256:abc123def456...",
								"is_forced":           false,
								"title":               "版本 v1.2.3 发布",
								"description":         "修复了一些重要bug",
								"release_notes":       "## 更新内容\n\n- 修复登录问题\n- 优化性能",
								"min_upgrade_version": "v1.0.0",
							},
						},
					},
					"400": map[string]interface{}{
						"description": "请求错误",
						"example": map[string]interface{}{
							"code":    400,
							"message": "Invalid request format",
							"error":   "missing required field: current_version",
						},
					},
					"401": map[string]interface{}{
						"description": "认证失败",
						"example": map[string]interface{}{
							"code":    401,
							"message": "Invalid credentials",
						},
					},
					"403": map[string]interface{}{
						"description": "权限不足",
						"example": map[string]interface{}{
							"code":    403,
							"message": "Insufficient permissions",
						},
					},
				},
			},
			{
				"name":        "下载开始记录",
				"method":      "POST",
				"path":        "/download-started",
				"permission":  "download",
				"description": "记录客户端开始下载新版本",
				"request": map[string]interface{}{
					"headers": map[string]string{
						"Authorization": "Bearer <app_id>:<api_key>",
						"Content-Type":  "application/json",
					},
					"body": map[string]interface{}{
						"version":   "string (required) - 下载的版本号",
						"client_id": "string (required) - 客户端唯一标识",
					},
					"example": map[string]interface{}{
						"version":   "v1.2.3",
						"client_id": "client_unique_id_12345",
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "记录成功",
						"example": map[string]interface{}{
							"code":    200,
							"message": "success",
							"data": map[string]interface{}{
								"message": "Download started recorded",
							},
						},
					},
				},
			},
			{
				"name":        "安装结果记录",
				"method":      "POST",
				"path":        "/install-result",
				"permission":  "install",
				"description": "记录客户端安装新版本的结果",
				"request": map[string]interface{}{
					"headers": map[string]string{
						"Authorization": "Bearer <app_id>:<api_key>",
						"Content-Type":  "application/json",
					},
					"body": map[string]interface{}{
						"version":       "string (required) - 安装的版本号",
						"client_id":     "string (required) - 客户端唯一标识",
						"success":       "boolean (required) - 安装是否成功",
						"error_message": "string (optional) - 失败时的错误信息",
					},
					"example_success": map[string]interface{}{
						"version":       "v1.2.3",
						"client_id":     "client_unique_id_12345",
						"success":       true,
						"error_message": "",
					},
					"example_failure": map[string]interface{}{
						"version":       "v1.2.3",
						"client_id":     "client_unique_id_12345",
						"success":       false,
						"error_message": "Checksum verification failed",
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "记录成功",
						"example": map[string]interface{}{
							"code":    200,
							"message": "success",
							"data": map[string]interface{}{
								"message": "Install result recorded",
							},
						},
					},
				},
			},
		},
		"examples": map[string]interface{}{
			"curl": map[string]interface{}{
				"check_update": `curl -X POST \
  'https://your-domain.com/api/v1/check-update' \
  -H 'Authorization: Bearer app_abc123:sk_def456789' \
  -H 'Content-Type: application/json' \
  -d '{
    "app_id": "app_abc123",
    "current_version": "v1.2.0",
    "channel": "stable",
    "client_id": "client_unique_id_12345",
    "region": "cn",
    "arch": "amd64",
    "os": "linux"
  }'`,
				"download_started": `curl -X POST \
  'https://your-domain.com/api/v1/download-started' \
  -H 'Authorization: Bearer app_abc123:sk_def456789' \
  -H 'Content-Type: application/json' \
  -d '{
    "version": "v1.2.3",
    "client_id": "client_unique_id_12345"
  }'`,
				"install_result": `curl -X POST \
  'https://your-domain.com/api/v1/install-result' \
  -H 'Authorization: Bearer app_abc123:sk_def456789' \
  -H 'Content-Type: application/json' \
  -d '{
    "version": "v1.2.3",
    "client_id": "client_unique_id_12345",
    "success": true,
    "error_message": ""
  }'`,
			},
			"javascript": map[string]interface{}{
				"check_update": `const response = await fetch('/api/v1/check-update', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer app_abc123:sk_def456789',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    app_id: 'app_abc123',
    current_version: 'v1.2.0',
    channel: 'stable',
    client_id: 'client_unique_id_12345',
    region: 'cn',
    arch: 'amd64',
    os: 'linux'
  })
});

const data = await response.json();
console.log(data);`,
				"error_handling": `try {
  const response = await fetch('/api/v1/check-update', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer app_abc123:sk_def456789',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(requestData)
  });

  if (!response.ok) {
    throw new Error(` + "`HTTP error! status: ${response.status}`" + `);
  }

  const data = await response.json();
  
  if (data.code !== 200) {
    throw new Error(` + "`API error: ${data.message}`" + `);
  }

  // 处理成功响应
  if (data.data.has_update) {
    // 有更新可用
    console.log('New version available:', data.data.latest_version);
  } else {
    // 已是最新版本
    console.log('Already up to date');
  }
} catch (error) {
  console.error('Update check failed:', error);
}`,
			},
			"go": map[string]interface{}{
				"check_update": `package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type CheckUpdateRequest struct {
    AppID          string ` + "`json:\"app_id\"`" + `
    CurrentVersion string ` + "`json:\"current_version\"`" + `
    Channel        string ` + "`json:\"channel\"`" + `
    ClientID       string ` + "`json:\"client_id\"`" + `
    Region         string ` + "`json:\"region,omitempty\"`" + `
    Arch           string ` + "`json:\"arch,omitempty\"`" + `
    OS             string ` + "`json:\"os,omitempty\"`" + `
}

type CheckUpdateResponse struct {
    Code    int                 ` + "`json:\"code\"`" + `
    Message string              ` + "`json:\"message\"`" + `
    Data    CheckUpdateData     ` + "`json:\"data\"`" + `
}

type CheckUpdateData struct {
    HasUpdate         bool   ` + "`json:\"has_update\"`" + `
    LatestVersion     string ` + "`json:\"latest_version,omitempty\"`" + `
    DownloadURL       string ` + "`json:\"download_url,omitempty\"`" + `
    FileSize          int64  ` + "`json:\"file_size,omitempty\"`" + `
    FileChecksum      string ` + "`json:\"file_checksum,omitempty\"`" + `
    IsForced          bool   ` + "`json:\"is_forced,omitempty\"`" + `
    Title             string ` + "`json:\"title,omitempty\"`" + `
    Description       string ` + "`json:\"description,omitempty\"`" + `
    ReleaseNotes      string ` + "`json:\"release_notes,omitempty\"`" + `
    MinUpgradeVersion string ` + "`json:\"min_upgrade_version,omitempty\"`" + `
}

func checkUpdate(appID, apiKey, currentVersion string) (*CheckUpdateResponse, error) {
    req := CheckUpdateRequest{
        AppID:          appID,
        CurrentVersion: currentVersion,
        Channel:        "stable",
        ClientID:       "client_unique_id_12345",
        Region:         "cn",
        Arch:           "amd64",
        OS:             "linux",
    }

    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    httpReq, err := http.NewRequest("POST", "https://your-domain.com/api/v1/check-update", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s:%s", appID, apiKey))
    httpReq.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result CheckUpdateResponse
    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
        return nil, err
    }

    return &result, nil
}`,
			},
		},
		"best_practices": []map[string]interface{}{
			{
				"title":   "API密钥安全",
				"content": "API密钥包含敏感信息，请妥善保管：\n1. 不要在客户端代码中硬编码API密钥\n2. 使用环境变量或配置文件存储密钥\n3. 定期轮换API密钥\n4. 为不同环境（开发、测试、生产）使用不同的密钥",
			},
			{
				"title":   "错误处理",
				"content": "正确处理API错误响应：\n1. 检查HTTP状态码\n2. 解析响应中的错误信息\n3. 实现重试机制（针对临时性错误）\n4. 记录错误日志便于调试",
			},
			{
				"title":   "版本比较",
				"content": "实现正确的版本比较逻辑：\n1. 使用语义化版本号（SemVer）\n2. 正确处理预发布版本（如v1.0.0-beta.1）\n3. 考虑兼容性和最小升级版本要求",
			},
			{
				"title":   "更新流程",
				"content": "推荐的更新流程：\n1. 检查更新 (check-update)\n2. 下载文件并验证校验和\n3. 记录下载开始 (download-started)\n4. 安装更新\n5. 记录安装结果 (install-result)\n6. 重启应用（如需要）",
			},
		},
	}

	c.JSON(http.StatusOK, models.SuccessResponse(docs))
}
