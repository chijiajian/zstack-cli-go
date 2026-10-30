// Package waiter polls observations without depending on Cobra, HTTP, or SDK types.
package waiter

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Observation struct {
	ResourceType string `json:"resourceType" yaml:"resourceType"`
	UUID         string `json:"uuid" yaml:"uuid"`
	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
	State        string `json:"state,omitempty" yaml:"state,omitempty"`
	Status       string `json:"status,omitempty" yaml:"status,omitempty"`
	JobResult    string `json:"jobResult,omitempty" yaml:"jobResult,omitempty"`
}

func (o Observation) Summary() string {
	parts := []string{}
	if o.State != "" {
		parts = append(parts, "state="+o.State)
	}
	if o.Status != "" {
		parts = append(parts, "status="+o.Status)
	}
	if len(parts) == 0 {
		return "state/status unavailable"
	}
	return strings.Join(parts, ", ")
}

type Observe func(context.Context) (Observation, error)
type Evaluate func(Observation, string) (bool, error)

func Match(o Observation, target string) (bool, error) {
	return strings.EqualFold(o.State, target) || strings.EqualFold(o.Status, target), nil
}

// Long-job terminal states follow ZStack's LongJobState contract. An explicitly
// requested terminal state is a successful wait, even when that state is Failed.
func Job(o Observation, target string) (bool, error) {
	if ok, _ := Match(o, target); ok {
		return true, nil
	}
	switch strings.ToLower(o.State) {
	case "succeeded", "failed", "canceled":
		return false, fmt.Errorf("job reached terminal state %s; jobResult=%s", o.State, o.JobResult)
	}
	return false, nil
}

// Run evaluates the initial observation before polling and preserves it on errors.
// notify is called only when the observation changes and must not block.
func Run(ctx context.Context, initial Observation, target string, interval time.Duration, observe Observe, evaluate Evaluate, notify func(Observation) error) (Observation, error) {
	target = strings.TrimSpace(target)
	if target == "" || interval <= 0 {
		return initial, fmt.Errorf("target must be nonempty and interval must be positive")
	}
	last := initial
	var previous Observation
	notified := false
	fail := func(err error) (Observation, error) {
		return last, fmt.Errorf("waiting for %s/%s to reach %q; last observed %s: %w", last.ResourceType, last.UUID, target, last.Summary(), err)
	}
	for {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		if notify != nil && (!notified || last != previous) {
			if err := notify(last); err != nil {
				return fail(err)
			}
			if err := ctx.Err(); err != nil {
				return fail(err)
			}
			previous = last
			notified = true
		}
		done, err := evaluate(last, target)
		if err != nil {
			return fail(err)
		}
		if done {
			return last, nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fail(ctx.Err())
		case <-timer.C:
		}
		next, err := observe(ctx)
		if err != nil {
			return fail(err)
		}
		last = next
	}
}
