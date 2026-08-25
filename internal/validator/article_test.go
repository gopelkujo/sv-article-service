package validator_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gopelkujo/sv-article-service/internal/validator"
)

func validInput() validator.ArticleInput {
	return validator.ArticleInput{
		Title:    "A valid article title!",
		Content:  strings.Repeat("c", 200),
		Category: "tech",
		Status:   "draft",
	}
}

func TestValidateArticle_Success(t *testing.T) {
	t.Parallel()
	require.NoError(t, validator.ValidateArticle(validInput()))
}

func TestValidateArticle_AllFieldsInvalid(t *testing.T) {
	t.Parallel()

	err := validator.ValidateArticle(validator.ArticleInput{
		Title:    "short",
		Content:  "too short",
		Category: "ab",
		Status:   "archived",
	})
	require.Error(t, err)

	ve, ok := validator.AsValidationError(err)
	require.True(t, ok)
	require.Equal(t, "validation failed", ve.Message)
	require.Len(t, ve.Details, 4)

	fields := map[string]string{}
	for _, d := range ve.Details {
		fields[d.Field] = d.Message
	}
	require.Equal(t, "must be at least 20 characters", fields["title"])
	require.Equal(t, "must be at least 200 characters", fields["content"])
	require.Equal(t, "must be at least 3 characters", fields["category"])
	require.Equal(t, "must be one of publish, draft, thrash", fields["status"])
}

func TestValidateArticle_RequiredFields(t *testing.T) {
	t.Parallel()

	err := validator.ValidateArticle(validator.ArticleInput{})
	ve, ok := validator.AsValidationError(err)
	require.True(t, ok)
	require.Len(t, ve.Details, 4)

	for _, d := range ve.Details {
		require.Equal(t, "is required", d.Message)
	}
}

func TestValidateArticle_AllowedStatuses(t *testing.T) {
	t.Parallel()

	for _, status := range []string{"publish", "draft", "thrash"} {
		input := validInput()
		input.Status = status
		require.NoError(t, validator.ValidateArticle(input), status)
	}
}

func TestValidatePagination(t *testing.T) {
	t.Parallel()

	require.NoError(t, validator.ValidatePagination(10, 0))
	require.NoError(t, validator.ValidatePagination(1, 100))

	err := validator.ValidatePagination(0, -1)
	ve, ok := validator.AsValidationError(err)
	require.True(t, ok)
	require.Len(t, ve.Details, 2)
}

func TestValidateID(t *testing.T) {
	t.Parallel()

	require.NoError(t, validator.ValidateID(1))
	require.Error(t, validator.ValidateID(0))
	require.Error(t, validator.ValidateID(-5))
}
