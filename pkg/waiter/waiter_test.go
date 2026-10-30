package waiter

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestJobTerminalPolicy(t *testing.T) {
	for _, state := range []string{"Succeeded", "Failed", "Canceled", "Waiting", "Suspended", "Running", "Canceling"} {
		o := Observation{State: state, JobResult: "backend failure"}
		done, err := Job(o, "Succeeded")
		if done != (state == "Succeeded") || (err != nil) != (state == "Failed" || state == "Canceled") {
			t.Fatalf("%s: %v %v", state, done, err)
		}
		done, err = Job(o, strings.ToLower(state))
		if !done || err != nil {
			t.Fatalf("explicit target %s: %v %v", state, done, err)
		}
	}
}

func TestRunUsesInitialObservationAndReportsOnlyChanges(t *testing.T) {
	initial := Observation{ResourceType: "long-job", UUID: "id", State: "Running"}
	calls, notifications := 0, 0
	result, err := Run(context.Background(), initial, "Succeeded", time.Microsecond, func(context.Context) (Observation, error) {
		calls++
		if calls == 2 {
			initial.State = "Succeeded"
		}
		return initial, nil
	}, Job, func(Observation) error { notifications++; return nil })
	if err != nil || result.State != "Succeeded" || calls != 2 || notifications != 2 {
		t.Fatalf("%+v %v calls=%d notify=%d", result, err, calls, notifications)
	}
	_, err = Run(context.Background(), result, "Succeeded", time.Hour, func(context.Context) (Observation, error) {
		t.Fatal("already satisfied must not poll")
		return Observation{}, nil
	}, Job, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunPreservesLastObservationOnFailureAndCancellation(t *testing.T) {
	initial := Observation{ResourceType: "instance", UUID: "id", State: "Stopped"}
	sentinel := errors.New("read failed")
	result, err := Run(context.Background(), initial, "Running", time.Microsecond, func(context.Context) (Observation, error) { return Observation{}, sentinel }, Match, nil)
	if !errors.Is(err, sentinel) || result != initial || !strings.Contains(err.Error(), "state=Stopped") {
		t.Fatalf("%+v %v", result, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	_, err = Run(ctx, initial, "Running", time.Hour, func(context.Context) (Observation, error) {
		t.Fatal("canceled wait must not poll")
		return Observation{}, nil
	}, Match, func(Observation) error { cancel(); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestNotificationCancellationBeatsReachedTarget(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initial := Observation{State: "Running"}
	_, err := Run(ctx, initial, "Running", time.Second, func(context.Context) (Observation, error) {
		t.Fatal("unexpected poll")
		return Observation{}, nil
	}, Match, func(Observation) error { cancel(); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled wait succeeded: %v", err)
	}
}
