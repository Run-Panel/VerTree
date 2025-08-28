package i18n

// Error message constants for internationalization support
// These constants provide consistent error messages that can be localized
const (
	// Application related messages
	ErrApplicationIDRequired   = "application_id_required"
	ErrApplicationNotFound     = "application_not_found"
	ErrApplicationCreateFailed = "application_create_failed"
	ErrApplicationUpdateFailed = "application_update_failed"
	ErrApplicationDeleteFailed = "application_delete_failed"
	ErrApplicationFetchFailed  = "application_fetch_failed"
	ErrInvalidRequestFormat    = "invalid_request_format"

	// API Key related messages
	ErrAPIKeyIDRequired      = "api_key_id_required"
	ErrAPIKeyNotFound        = "api_key_not_found"
	ErrAPIKeyCreateFailed    = "api_key_create_failed"
	ErrAPIKeyUpdateFailed    = "api_key_update_failed"
	ErrAPIKeyDeleteFailed    = "api_key_delete_failed"
	ErrAPIKeyFetchFailed     = "api_key_fetch_failed"
	ErrAPIKeyNameExists      = "api_key_name_exists"
	ErrAPIKeySecretGenFailed = "api_key_secret_generation_failed"

	// Authentication related messages
	ErrAdminIDNotFound = "admin_id_not_found"
	ErrUnauthorized    = "unauthorized"

	// Success messages
	MsgApplicationDeleted = "application_deleted_successfully"
	MsgAPIKeyDeleted      = "api_key_deleted_successfully"
)

// Default English messages
var EnglishMessages = map[string]string{
	ErrApplicationIDRequired:   "Application ID is required",
	ErrApplicationNotFound:     "Application not found",
	ErrApplicationCreateFailed: "Failed to create application",
	ErrApplicationUpdateFailed: "Failed to update application",
	ErrApplicationDeleteFailed: "Failed to delete application",
	ErrApplicationFetchFailed:  "Failed to fetch applications",
	ErrInvalidRequestFormat:    "Invalid request format",

	ErrAPIKeyIDRequired:      "Application ID and Key ID are required",
	ErrAPIKeyNotFound:        "Application key not found",
	ErrAPIKeyCreateFailed:    "Failed to create application key",
	ErrAPIKeyUpdateFailed:    "Failed to update application key",
	ErrAPIKeyDeleteFailed:    "Failed to delete application key",
	ErrAPIKeyFetchFailed:     "Failed to fetch application keys",
	ErrAPIKeyNameExists:      "Key with this name already exists for this application",
	ErrAPIKeySecretGenFailed: "Failed to generate key secret",

	ErrAdminIDNotFound: "Admin ID not found in context",
	ErrUnauthorized:    "Unauthorized",

	MsgApplicationDeleted: "Application deleted successfully",
	MsgAPIKeyDeleted:      "Application key deleted successfully",
}

// Chinese messages
var ChineseMessages = map[string]string{
	ErrApplicationIDRequired:   "应用程序ID是必需的",
	ErrApplicationNotFound:     "未找到应用程序",
	ErrApplicationCreateFailed: "创建应用程序失败",
	ErrApplicationUpdateFailed: "更新应用程序失败",
	ErrApplicationDeleteFailed: "删除应用程序失败",
	ErrApplicationFetchFailed:  "获取应用程序失败",
	ErrInvalidRequestFormat:    "无效的请求格式",

	ErrAPIKeyIDRequired:      "应用程序ID和密钥ID是必需的",
	ErrAPIKeyNotFound:        "未找到应用程序密钥",
	ErrAPIKeyCreateFailed:    "创建应用程序密钥失败",
	ErrAPIKeyUpdateFailed:    "更新应用程序密钥失败",
	ErrAPIKeyDeleteFailed:    "删除应用程序密钥失败",
	ErrAPIKeyFetchFailed:     "获取应用程序密钥失败",
	ErrAPIKeyNameExists:      "此应用程序已存在同名密钥",
	ErrAPIKeySecretGenFailed: "生成密钥失败",

	ErrAdminIDNotFound: "在上下文中未找到管理员ID",
	ErrUnauthorized:    "未授权",

	MsgApplicationDeleted: "应用程序删除成功",
	MsgAPIKeyDeleted:      "应用程序密钥删除成功",
}

// Localizer interface for message localization
type Localizer interface {
	Get(key string) string
}

// SimpleLocalizer provides basic localization functionality
type SimpleLocalizer struct {
	messages map[string]string
}

// NewLocalizer creates a new localizer with the specified language
func NewLocalizer(lang string) Localizer {
	switch lang {
	case "zh", "zh-CN", "chinese":
		return &SimpleLocalizer{messages: ChineseMessages}
	default:
		return &SimpleLocalizer{messages: EnglishMessages}
	}
}

// Get retrieves a localized message by key
func (l *SimpleLocalizer) Get(key string) string {
	if msg, exists := l.messages[key]; exists {
		return msg
	}
	// Fallback to English if key not found
	if msg, exists := EnglishMessages[key]; exists {
		return msg
	}
	// Ultimate fallback
	return key
}
