package handler

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
)

// CustomValidator wraps go-playground/validator for Echo.
type CustomValidator struct {
	validate *validator.Validate
}

// NewValidator returns an Echo-compatible struct validator.
func NewValidator() *CustomValidator {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(jsonFieldName)
	return &CustomValidator{validate: v}
}

// Validate implements echo.Validator.
func (cv *CustomValidator) Validate(i any) error {
	return cv.validate.Struct(i)
}

// BindAndValidate binds the request into dest and validates it.
func BindAndValidate(c echo.Context, dest any) error {
	if err := c.Bind(dest); err != nil {
		return domain.NewValidationError("Invalid request body", map[string]string{
			"body": "Malformed JSON or invalid parameter format",
		})
	}
	if err := c.Validate(dest); err != nil {
		return toValidationError(err)
	}
	return nil
}

func toValidationError(err error) error {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return domain.NewValidationError("Validation failed", map[string]string{
			"request": "Invalid input",
		})
	}

	fields := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		fields[fe.Field()] = validationMessage(fe)
	}

	return domain.NewValidationError("Validation failed", fields)
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + fe.Param() + " characters"
	case "max":
		return "Must be at most " + fe.Param() + " characters"
	case "gt":
		return "Must be greater than " + fe.Param()
	case "oneof":
		return "Must be one of: " + fe.Param()
	default:
		return "Invalid value"
	}
}

func jsonFieldName(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "" {
		return fld.Name
	}
	return name
}
