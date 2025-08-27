package middleware

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom tag name function
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidateJSON validates JSON request body against a struct
func ValidateJSON(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(obj); err != nil {
			c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid JSON format", err))
			c.Abort()
			return
		}

		if err := validate.Struct(obj); err != nil {
			var errorMessages []string
			for _, err := range err.(validator.ValidationErrors) {
				errorMessages = append(errorMessages, formatValidationError(err))
			}

			c.JSON(http.StatusBadRequest, models.BadRequestResponse(
				"Validation failed: "+strings.Join(errorMessages, ", "),
				nil,
			))
			c.Abort()
			return
		}

		c.Set("validatedData", obj)
		c.Next()
	}
}

// ValidateQuery validates query parameters against a struct
func ValidateQuery(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindQuery(obj); err != nil {
			c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid query parameters", err))
			c.Abort()
			return
		}

		if err := validate.Struct(obj); err != nil {
			var errorMessages []string
			for _, err := range err.(validator.ValidationErrors) {
				errorMessages = append(errorMessages, formatValidationError(err))
			}

			c.JSON(http.StatusBadRequest, models.BadRequestResponse(
				"Validation failed: "+strings.Join(errorMessages, ", "),
				nil,
			))
			c.Abort()
			return
		}

		c.Set("validatedQuery", obj)
		c.Next()
	}
}

// formatValidationError formats a validation error into a human-readable message
func formatValidationError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, err.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, err.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
