package spotify

import (
	"context"
	"errors"
	"testing"
)

type mockPaging struct {
	Items []int
	Next  string
}

func TestCollectPaged_InvalidMax(t *testing.T) {
	_, err := collectPaged(context.Background(), 10, -1, func(ctx context.Context, limit, offset int) (paging[int], error) {
		return paging[int]{}, nil
	})
	if err == nil || err.Error() != "max must be >= 0" {
		t.Errorf("expected 'max must be >= 0' error, got: %v", err)
	}
}

func TestCollectPaged_EmptyResult(t *testing.T) {
	callCount := 0
	result, err := collectPaged(context.Background(), 10, 0, func(ctx context.Context, limit, offset int) (paging[int], error) {
		callCount++
		return paging[int]{Items: []int{}, Next: ""}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d items", len(result))
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestCollectPaged_SinglePage(t *testing.T) {
	result, err := collectPaged(context.Background(), 10, 0, func(ctx context.Context, limit, offset int) (paging[int], error) {
		if offset != 0 {
			t.Errorf("unexpected offset: %d", offset)
		}
		return paging[int]{Items: []int{1, 2, 3}, Next: ""}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}
	if result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("unexpected items: %v", result)
	}
}

func TestCollectPaged_MultiplePages(t *testing.T) {
	callCount := 0
	result, err := collectPaged(context.Background(), 2, 0, func(ctx context.Context, limit, offset int) (paging[int], error) {
		callCount++
		switch offset {
		case 0:
			if limit != 2 {
				t.Errorf("expected limit 2, got %d", limit)
			}
			return paging[int]{Items: []int{1, 2}, Next: "page2"}, nil
		case 2:
			return paging[int]{Items: []int{3, 4}, Next: "page3"}, nil
		case 4:
			return paging[int]{Items: []int{5}, Next: ""}, nil
		default:
			return paging[int]{}, errors.New("unexpected offset")
		}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("expected 5 items, got %d", len(result))
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestCollectPaged_WithMaxLimit(t *testing.T) {
	callCount := 0
	result, err := collectPaged(context.Background(), 10, 3, func(ctx context.Context, limit, offset int) (paging[int], error) {
		callCount++
		if offset == 0 {
			if limit != 3 { // max is 3, should limit first page
				t.Errorf("expected limit 3, got %d", limit)
			}
			return paging[int]{Items: []int{1, 2, 3}, Next: "page2"}, nil
		}
		t.Error("should not fetch second page when max is reached")
		return paging[int]{}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestCollectPaged_MaxLimitAcrossPages(t *testing.T) {
	callCount := 0
	result, err := collectPaged(context.Background(), 10, 15, func(ctx context.Context, limit, offset int) (paging[int], error) {
		callCount++
		switch offset {
		case 0:
			if limit != 10 {
				t.Errorf("expected limit 10, got %d", limit)
			}
			return paging[int]{Items: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, Next: "page2"}, nil
		case 10:
			if limit != 5 { // remaining = 15 - 10 = 5
				t.Errorf("expected limit 5, got %d", limit)
			}
			return paging[int]{Items: []int{11, 12, 13, 14, 15, 16, 17}, Next: "page3"}, nil
		default:
			t.Error("should not fetch third page")
			return paging[int]{}, nil
		}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 15 {
		t.Errorf("expected 15 items, got %d", len(result))
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestCollectPaged_ErrorHandling(t *testing.T) {
	expectedErr := errors.New("fetch failed")
	_, err := collectPaged(context.Background(), 10, 0, func(ctx context.Context, limit, offset int) (paging[int], error) {
		if offset == 0 {
			return paging[int]{}, expectedErr
		}
		return paging[int]{}, nil
	})
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestCollectPaged_EmptyNextStopsPaging(t *testing.T) {
	callCount := 0
	result, err := collectPaged(context.Background(), 10, 0, func(ctx context.Context, limit, offset int) (paging[int], error) {
		callCount++
		if offset == 0 {
			return paging[int]{Items: []int{1, 2}, Next: ""}, nil // empty Next stops pagination
		}
		t.Error("should not be called again")
		return paging[int]{}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestCollectPaged_EmptyItemsStopsPaging(t *testing.T) {
	callCount := 0
	result, err := collectPaged(context.Background(), 10, 0, func(ctx context.Context, limit, offset int) (paging[int], error) {
		callCount++
		if offset == 0 {
			return paging[int]{Items: []int{}, Next: "hasNext"}, nil // empty items stops pagination
		}
		t.Error("should not be called again")
		return paging[int]{}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 items, got %d", len(result))
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}
