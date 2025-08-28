package models

// APIResponse represents a generic API response
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationRequest represents pagination parameters in requests
type PaginationRequest struct {
	Page  int `json:"page" form:"page" validate:"min=1"`
	Limit int `json:"limit" form:"limit" validate:"min=1,max=100"`
}

// PaginationResponse represents pagination information in responses
type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Code       int                `json:"code"`
	Message    string             `json:"message"`
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// CheckUpdateRequest represents the client update check request
type CheckUpdateRequest struct {
	AppID          string `json:"app_id" validate:"required"` // Required for app-scoped updates
	CurrentVersion string `json:"current_version" validate:"required"`
	Channel        string `json:"channel" validate:"required,oneof=stable beta alpha"`
	ClientID       string `json:"client_id" validate:"required"`
	Region         string `json:"region"`
	Arch           string `json:"arch"`
	OS             string `json:"os"`
}

// CheckUpdateResponse represents the client update check response
type CheckUpdateResponse struct {
	HasUpdate         bool   `json:"has_update"`
	LatestVersion     string `json:"latest_version,omitempty"`
	DownloadURL       string `json:"download_url,omitempty"`
	FileSize          int64  `json:"file_size,omitempty"`
	FileChecksum      string `json:"file_checksum,omitempty"`
	IsForced          bool   `json:"is_forced,omitempty"`
	Title             string `json:"title,omitempty"`
	Description       string `json:"description,omitempty"`
	ReleaseNotes      string `json:"release_notes,omitempty"`
	MinUpgradeVersion string `json:"min_upgrade_version,omitempty"`
}

// StatsRequest represents the statistics query request
type StatsRequest struct {
	Period string `json:"period" form:"period" validate:"oneof=1d 7d 30d 90d"`
	Action string `json:"action" form:"action" validate:"oneof=all check download install success failed"`
}

// StatsResponse represents the statistics response
type StatsResponse struct {
	TotalUsers          int64            `json:"totalUsers"`
	TotalDownloads      int64            `json:"totalDownloads"`
	SuccessRate         float64          `json:"successRate"`
	VersionDistribution map[string]int64 `json:"versionDistribution"`
	RegionDistribution  map[string]int64 `json:"regionDistribution"`
	DailyStats          []DailyStat      `json:"dailyStats"`
}

// DailyStat represents daily statistics
type DailyStat struct {
	Date      string `json:"date"`
	Downloads int64  `json:"downloads"`
	Installs  int64  `json:"installs"`
	Failures  int64  `json:"failures"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// SuccessResponse creates a success API response
func SuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// SuccessResponseWithMessage creates a success API response with custom message
func SuccessResponseWithMessage(message string, data interface{}) *APIResponse {
	return &APIResponse{
		Code:    200,
		Message: message,
		Data:    data,
	}
}

// ErrorResponseWithCode creates an error API response with custom code
func ErrorResponseWithCode(code int, message string, err error) *ErrorResponse {
	response := &ErrorResponse{
		Code:    code,
		Message: message,
	}
	if err != nil {
		response.Error = err.Error()
	}
	return response
}

// BadRequestResponse creates a bad request error response
func BadRequestResponse(message string, err error) *ErrorResponse {
	return ErrorResponseWithCode(400, message, err)
}

// NotFoundResponse creates a not found error response
func NotFoundResponse(message string) *ErrorResponse {
	return ErrorResponseWithCode(404, message, nil)
}

// InternalServerErrorResponse creates an internal server error response
func InternalServerErrorResponse(message string, err error) *ErrorResponse {
	return ErrorResponseWithCode(500, message, err)
}

// UnauthorizedResponse creates an unauthorized error response
func UnauthorizedResponse(message string) *ErrorResponse {
	return ErrorResponseWithCode(401, message, nil)
}

// ForbiddenResponse creates a forbidden error response
func ForbiddenResponse(message string) *ErrorResponse {
	return ErrorResponseWithCode(403, message, nil)
}

// ErrorResponseWithCodeAndError creates a generic error response with code, message and error code
func ErrorResponseWithCodeAndError(code int, message, errorCode string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Error:   errorCode,
	}
}
