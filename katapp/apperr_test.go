package katapp

import (
	"testing"
)

func TestNewErr(t *testing.T) {
	tests := []struct {
		name  string
		scope ErrScope
		msg   string
	}{
		{
			name:  "Unknown error",
			scope: ErrUnknown,
			msg:   "Unknown error occurred",
		},
		{
			name:  "Internal error",
			scope: ErrInternal,
			msg:   "Internal server error",
		},
		{
			name:  "Not found error",
			scope: ErrNotFound,
			msg:   "Resource not found",
		},
		{
			name:  "Invalid input error",
			scope: ErrInvalidInput,
			msg:   "Invalid input provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewErr(tt.scope, tt.msg)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}
			if err.Scope != tt.scope {
				t.Errorf("Expected scope %v, got %v", tt.scope, err.Scope)
			}
			if err.Msg != tt.msg {
				t.Errorf("Expected message %q, got %q", tt.msg, err.Msg)
			}
		})
	}
}

func TestErr_Error(t *testing.T) {
	err := &Err{
		Scope: ErrInternal,
		Msg:   "Internal error",
	}

	expected := "Internal error"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}
