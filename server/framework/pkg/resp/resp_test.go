package resp_test

import (
	"errors"
	"fmt"
	"testing"

	"gx1727.com/xin/framework/pkg/resp"
)

var ErrNotFound = resp.Err(2001, "not found")
var ErrPermission = resp.Err(4001, "no permission")

func TestBizError_Is_byCode(t *testing.T) {
	// Two different instances with the same Code should match.
	err1 := resp.Err(2001, "user not found")
	err2 := resp.Err(2001, "order not found")
	if !errors.Is(err1, err2) {
		t.Fatal("errors.Is should match by Code, not pointer identity")
	}
}

func TestBizError_Is_differentCode(t *testing.T) {
	if errors.Is(ErrNotFound, ErrPermission) {
		t.Fatal("different codes should not match")
	}
}

func TestBizError_Is_fmtErrorfWrap(t *testing.T) {
	// Sentinel wrapped with fmt.Errorf("%w") should still be detectable.
	wrapped := fmt.Errorf("extra context: %w", ErrNotFound)
	if !errors.Is(wrapped, ErrNotFound) {
		t.Fatal("sentinel wrapped with fmt.Errorf should be matched")
	}
}

func TestBizError_As(t *testing.T) {
	wrapped := fmt.Errorf("context: %w", ErrPermission)
	var bizErr *resp.BizError
	if !errors.As(wrapped, &bizErr) {
		t.Fatal("errors.As should extract BizError from wrapped error")
	}
	if bizErr.Code != 4001 {
		t.Fatalf("expected code 4001, got %d", bizErr.Code)
	}
}
