package validator

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	v := validator.New()

	_ = v.RegisterValidation("datetime", validateDateTime)

	return &Validator{validate: v}
}

func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

func validateDateTime(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	
	matched, _ := regexp.MatchString(`^(0[1-9]|1[0-2])-\d{4}$`, dateStr)
	return matched
}

func FormatValidationError(err error) string {
	if err == nil {
		return ""
	}

	var messages []string
	
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			messages = append(messages, formatFieldError(e))
		}
	}

	if len(messages) == 0 {
		return err.Error()
	}

	return strings.Join(messages, "; ")
}

func formatFieldError(e validator.FieldError) string {
	field := e.Field()
	
	switch e.Tag() {
	case "required":
		return field + " is required"
	case "uuid":
		return field + " must be a valid UUID"
	case "min":
		return field + " must be at least " + e.Param()
	case "max":
		return field + " must be at most " + e.Param()
	case "datetime":
		return field + " must be in format MM-YYYY (e.g., 07-2025)"
	default:
		return field + " is invalid"
	}
}
