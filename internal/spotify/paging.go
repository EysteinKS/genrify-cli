package spotify

import (
	"context"
	"fmt"
)

func collectPaged[T any](ctx context.Context, pageSize, max int, fetch func(ctx context.Context, limit, offset int) (paging[T], error)) ([]T, error) {
	if max < 0 {
		return nil, fmt.Errorf("max must be >= 0")
	}
	limit := pageSize
	if max > 0 && max < limit {
		limit = max
	}

	var out []T
	offset := 0
	for {
		p, err := fetch(ctx, limit, offset)
		if err != nil {
			return nil, err
		}
		out = append(out, p.Items...)

		if max > 0 && len(out) >= max {
			out = out[:max]
			return out, nil
		}
		if p.Next == "" || len(p.Items) == 0 {
			return out, nil
		}
		offset += limit

		if max > 0 {
			remaining := max - len(out)
			if remaining < limit {
				limit = remaining
			}
		}
	}
}
