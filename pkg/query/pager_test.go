package query

import (
	"context"
	"errors"
	"net/url"
	"reflect"
	"strconv"
	"testing"
)

func TestPaginationPreservesFiltersAndAdvancesByReceivedRows(t *testing.T) {
	input := url.Values{"q": {"state=Enabled", "zoneUuid=z"}, "limit": {"3"}, "start": {"1"}, "sort": {"-name"}}
	var starts []int
	total := 5
	got, err := Run(context.Background(), input, true, func(ctx context.Context, q url.Values) (Result[int], error) {
		start, _ := strconv.Atoi(q.Get("start"))
		starts = append(starts, start)
		if !reflect.DeepEqual(q["q"], input["q"]) || q.Get("sort") != "-name" || q.Get("limit") != "3" {
			t.Fatalf("lost options: %v", q)
		}
		// Server caps pages below the requested size.
		return Result[int]{Items: []int{start, start + 1}, Total: &total}, nil
	})
	if err != nil || !reflect.DeepEqual(got.Items, []int{1, 2, 3, 4}) || !reflect.DeepEqual(starts, []int{1, 3}) || input.Get("start") != "1" {
		t.Fatalf("result=%v starts=%v err=%v input=%v", got, starts, err, input)
	}
}

func TestPaginationFailuresDoNotReturnPartialSuccess(t *testing.T) {
	for _, scenario := range []string{"page-error", "empty", "missing-total", "cancel"} {
		t.Run(scenario, func(t *testing.T) {
			calls := 0
			total := 4
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			got, err := Run(ctx, url.Values{}, true, func(ctx context.Context, q url.Values) (Result[int], error) {
				calls++
				if calls == 1 {
					if scenario == "cancel" {
						cancel()
					}
					return Result[int]{Items: []int{1}, Total: &total}, nil
				}
				switch scenario {
				case "page-error":
					return Result[int]{}, errors.New("failed page")
				case "empty":
					return Result[int]{Total: &total}, nil
				default:
					return Result[int]{Items: []int{2}}, nil
				}
			})
			if err == nil || len(got.Items) != 0 {
				t.Fatalf("partial success: %+v %v", got, err)
			}
		})
	}
}

func TestPaginationRejectsOffsetOverflow(t *testing.T) {
	maximum := int(^uint(0) >> 1)
	result, err := Run(context.Background(), url.Values{"start": {strconv.Itoa(maximum)}}, true, func(context.Context, url.Values) (Result[int], error) {
		return Result[int]{Items: []int{1}, Total: &maximum}, nil
	})
	if err == nil || len(result.Items) != 0 {
		t.Fatalf("overflow returned success: %+v %v", result, err)
	}
}

func TestCancellationAfterFetchNeverReturnsSuccess(t *testing.T) {
	for _, mode := range []string{"single", "all", "count"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			values := url.Values{}
			if mode == "count" {
				values.Set("count", "true")
			}
			total := 1
			result, err := Run(ctx, values, mode != "single", func(context.Context, url.Values) (Result[int], error) {
				cancel()
				return Result[int]{Items: []int{1}, Total: &total}, nil
			})
			if !errors.Is(err, context.Canceled) || len(result.Items) != 0 || result.Total != nil {
				t.Fatalf("canceled query returned result: %+v %v", result, err)
			}
		})
	}
}

func TestSinglePageErrorDiscardsPartialResult(t *testing.T) {
	expected := errors.New("failed decode")
	result, err := Run(context.Background(), url.Values{}, false, func(context.Context, url.Values) (Result[int], error) { return Result[int]{Items: []int{1}}, expected })
	if !errors.Is(err, expected) || len(result.Items) != 0 {
		t.Fatalf("partial result: %+v %v", result, err)
	}
}
