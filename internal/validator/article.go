// Package validator provides centralized request validation rules.
package validator

import "errors"

// FieldError describes a single field-level validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError aggregates one or more field-level validation failures.
type ValidationError struct {
	Message string
	Details []FieldError
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "validation failed"
}

// ArticleInput is the validated payload for create and update operations.
type ArticleInput struct {
	Title    string
	Content  string
	Category string
	Status   string
}

// ValidateArticle validates create/update article fields and returns all
// violations at once. A nil error means the input is valid.
func ValidateArticle(input ArticleInput) error {
	details := make([]FieldError, 0)

	if input.Title == "" {
		details = append(details, FieldError{Field: "title", Message: "is required"})
	} else if len(input.Title) < 20 {
		details = append(details, FieldError{Field: "title", Message: "must be at least 20 characters"})
	}

	if input.Content == "" {
		details = append(details, FieldError{Field: "content", Message: "is required"})
	} else if len(input.Content) < 200 {
		details = append(details, FieldError{Field: "content", Message: "must be at least 200 characters"})
	}

	if input.Category == "" {
		details = append(details, FieldError{Field: "category", Message: "is required"})
	} else if len(input.Category) < 3 {
		details = append(details, FieldError{Field: "category", Message: "must be at least 3 characters"})
	}

	if input.Status == "" {
		details = append(details, FieldError{Field: "status", Message: "is required"})
	} else if !isAllowedStatus(input.Status) {
		details = append(details, FieldError{
			Field:   "status",
			Message: "must be one of publish, draft, thrash",
		})
	}

	if len(details) == 0 {
		return nil
	}

	return &ValidationError{
		Message: "validation failed",
		Details: details,
	}
}

// ValidatePagination ensures limit is a positive integer and offset is non-negative.
func ValidatePagination(limit, offset int) error {
	details := make([]FieldError, 0)

	if limit <= 0 {
		details = append(details, FieldError{
			Field:   "limit",
			Message: "must be a positive integer",
		})
	}
	if offset < 0 {
		details = append(details, FieldError{
			Field:   "offset",
			Message: "must be a non-negative integer",
		})
	}

	if len(details) == 0 {
		return nil
	}

	return &ValidationError{
		Message: "validation failed",
		Details: details,
	}
}

// ValidateID ensures the article id path parameter is a positive integer.
func ValidateID(id int) error {
	if id <= 0 {
		return &ValidationError{
			Message: "validation failed",
			Details: []FieldError{{
				Field:   "id",
				Message: "must be a positive integer",
			}},
		}
	}
	return nil
}

// AsValidationError extracts a *ValidationError from err using errors.As.
func AsValidationError(err error) (*ValidationError, bool) {
	var ve *ValidationError
	if errors.As(err, &ve) {
		return ve, true
	}
	return nil, false
}

func isAllowedStatus(status string) bool {
	switch status {
	case "publish", "draft", "thrash":
		return true
	default:
		return false
	}
}
