package unit

import (
	"testing"

	"github.com/yfh-yun/moviepilot-go/pkg/validator"
)

type TestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required"`
	Age   int    `validate:"required"`
}

func TestStringValidator(t *testing.T) {
	// Test required validation
	v := &validator.StringValidator{
		Required: true,
	}

	err := v.Validate("")
	if err == nil {
		t.Error("Expected validation error for empty required field")
	}

	// Test min length validation
	v = &validator.StringValidator{
		Min: 5,
	}

	err = v.Validate("abc")
	if err == nil {
		t.Error("Expected validation error for string too short")
	}

	// Test max length validation
	v = &validator.StringValidator{
		Max: 5,
	}

	err = v.Validate("abcdef")
	if err == nil {
		t.Error("Expected validation error for string too long")
	}

	// Test pattern validation
	v = &validator.StringValidator{
		Pattern: `^[a-zA-Z0-9]+$`,
	}

	err = v.Validate("test123")
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}

	err = v.Validate("test-123")
	if err == nil {
		t.Error("Expected validation error for invalid pattern")
	}
}

func TestValidateStruct(t *testing.T) {
	// Test valid struct
	valid := &TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	err := validator.ValidateStruct(valid)
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}

	// Test invalid struct (missing required fields)
	invalid := &TestStruct{
		Name: "John Doe",
		// Email missing
		// Age missing
	}

	err = validator.ValidateStruct(invalid)
	if err == nil {
		t.Error("Expected validation error for missing required fields")
	}
}