// Package query owns pagination independently of commands, SDK types, and output.
package query

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type Result[T any] struct {
	Items []T  `json:"items" yaml:"items"`
	Total *int `json:"total,omitempty" yaml:"total,omitempty"`
}

type FetchPage[T any] func(context.Context, url.Values) (Result[T], error)

// Option configures resource-specific query behavior without coupling the pager to APIs.
type Option func(*options)
type options struct{ defaultSort string }

func WithDefaultSort(sort string) Option { return func(o *options) { o.defaultSort = sort } }

func Run[T any](ctx context.Context, values url.Values, all bool, fetch FetchPage[T], overrides ...Option) (Result[T], error) {
	if err := ctx.Err(); err != nil {
		return Result[T]{}, err
	}
	settings := options{defaultSort: "+uuid"}
	for _, apply := range overrides {
		apply(&settings)
	}
	q := url.Values{}
	for k, v := range values {
		q[k] = append([]string(nil), v...)
	}
	if !all || q.Get("count") == "true" {
		r, err := fetch(ctx, q)
		if err != nil {
			return Result[T]{}, err
		}
		if err := ctx.Err(); err != nil {
			return Result[T]{}, err
		}
		if r.Items == nil {
			r.Items = []T{}
		}
		return r, err
	}
	size, _ := strconv.Atoi(q.Get("limit"))
	if size <= 0 {
		size = 100
	}
	start, _ := strconv.Atoi(q.Get("start"))
	q.Set("limit", strconv.Itoa(size))
	q.Set("replyWithCount", "true")
	// Stable ordering reduces omissions when the server's default order is unspecified.
	if q.Get("sort") == "" {
		q.Set("sort", settings.defaultSort)
	}
	result := Result[T]{Items: []T{}}
	for {
		if err := ctx.Err(); err != nil {
			return Result[T]{}, err
		}
		q.Set("start", strconv.Itoa(start))
		page, err := fetch(ctx, q)
		if err != nil {
			return Result[T]{}, err
		}
		if err := ctx.Err(); err != nil {
			return Result[T]{}, err
		}
		if page.Total == nil {
			return Result[T]{}, fmt.Errorf("server omitted total for paginated query")
		}
		if *page.Total < 0 {
			return Result[T]{}, fmt.Errorf("invalid negative total")
		}
		result.Total = page.Total
		result.Items = append(result.Items, page.Items...)
		if len(page.Items) > int(^uint(0)>>1)-start {
			return Result[T]{}, fmt.Errorf("pagination offset overflow")
		}
		start += len(page.Items)
		if start >= *page.Total {
			return result, nil
		}
		if len(page.Items) == 0 {
			return Result[T]{}, fmt.Errorf("pagination stopped before total: received %d of %d", start, *page.Total)
		}
	}
}
