package spotify

import (
	"testing"
)

func TestDecodeAPIError_ValidJSON(t *testing.T) {
	body := []byte(`{"error": {"status": 404, "message": "Not found"}}`)
	err := decodeAPIError(body, 500)

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.Status != 404 {
		t.Errorf("expected status 404, got %d", apiErr.Status)
	}
	if apiErr.Message != "Not found" {
		t.Errorf("expected message 'Not found', got %q", apiErr.Message)
	}
}

func TestDecodeAPIError_MessageOnly(t *testing.T) {
	body := []byte(`{"error": {"message": "Bad request"}}`)
	err := decodeAPIError(body, 400)

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.Status != 400 {
		t.Errorf("expected status 400 (fallback), got %d", apiErr.Status)
	}
	if apiErr.Message != "Bad request" {
		t.Errorf("expected message 'Bad request', got %q", apiErr.Message)
	}
}

func TestDecodeAPIError_StatusOnly(t *testing.T) {
	body := []byte(`{"error": {"status": 429}}`)
	err := decodeAPIError(body, 500)

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.Status != 429 {
		t.Errorf("expected status 429, got %d", apiErr.Status)
	}
	if apiErr.Message != "" {
		t.Errorf("expected empty message, got %q", apiErr.Message)
	}
}

func TestDecodeAPIError_InvalidJSON(t *testing.T) {
	body := []byte(`invalid json`)
	err := decodeAPIError(body, 503)

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.Status != 503 {
		t.Errorf("expected status 503 (fallback), got %d", apiErr.Status)
	}
	if apiErr.Message != "" {
		t.Errorf("expected empty message, got %q", apiErr.Message)
	}
}

func TestDecodeAPIError_EmptyError(t *testing.T) {
	body := []byte(`{"error": {}}`)
	err := decodeAPIError(body, 500)

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.Status != 500 {
		t.Errorf("expected status 500 (fallback), got %d", apiErr.Status)
	}
}

func TestDecodeAPIError_NoError(t *testing.T) {
	body := []byte(`{}`)
	err := decodeAPIError(body, 500)

	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.Status != 500 {
		t.Errorf("expected status 500 (fallback), got %d", apiErr.Status)
	}
}

func TestAPIError_Error_WithMessage(t *testing.T) {
	err := APIError{Status: 401, Message: "Invalid token"}
	expected := "spotify api error: http 401: Invalid token"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestAPIError_Error_WithoutMessage(t *testing.T) {
	err := APIError{Status: 500}
	expected := "spotify api error: http 500"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}
