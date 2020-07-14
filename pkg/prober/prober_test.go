package prober

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/radhus/voff/pkg/config"

	"github.com/stretchr/testify/assert"
)

var errFakeFailed = errors.New("fake failed immediately")

func fakeFailImmediately(context.Context) error {
	return errFakeFailed
}

func fakeSucceedImmediately(context.Context) error {
	return nil
}

func withDelay(delay time.Duration, inner probeFunc) probeFunc {
	return func(ctx context.Context) error {
		time.Sleep(delay)
		return inner(ctx)
	}
}

type ctxKey int

var sequenceKey ctxKey

func withSequence(sequence *[]probeFunc) (context.Context, probeFunc) {
	ctx := context.WithValue(context.Background(), sequenceKey, sequence)
	return ctx, func(ctx context.Context) error {
		sequence := ctx.Value(sequenceKey).(*[]probeFunc)
		pop := (*sequence)[0]
		*sequence = (*sequence)[1:]
		return pop(ctx)
	}
}

func newWithProbeFunc(probeConfig *config.Probe, probeFunc probeFunc) Prober {
	p := New(probeConfig).(*prober)
	p.probe = probeFunc
	return p
}

func TestProbe_WithSuccessProbe(t *testing.T) {
	p := newWithProbeFunc(&config.Probe{TimeoutSeconds: 3}, fakeSucceedImmediately)
	err := p.Probe(context.Background())
	assert.Nil(t, err)
}

func TestProbe_WithFailureProbe(t *testing.T) {
	p := newWithProbeFunc(&config.Probe{TimeoutSeconds: 3}, fakeFailImmediately)
	err := p.Probe(context.Background())
	assert.Equal(t, errFakeFailed, err)
}

func TestProbe_WhenTimeout(t *testing.T) {
	p := newWithProbeFunc(&config.Probe{TimeoutSeconds: 1}, withDelay(2*time.Second, fakeSucceedImmediately))
	err := p.Probe(context.Background())
	assert.NotNil(t, err)
	assert.Contains(t, "context deadline exceeded", err.Error())
}

func TestProbe_WhenSlow(t *testing.T) {
	p := newWithProbeFunc(&config.Probe{TimeoutSeconds: 2}, withDelay(1*time.Second, fakeSucceedImmediately))
	err := p.Probe(context.Background())
	assert.Nil(t, err)
}

func TestProbe_WhenCancelled(t *testing.T) {
	p := newWithProbeFunc(&config.Probe{TimeoutSeconds: 60}, withDelay(30*time.Second, fakeSucceedImmediately))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		errCh <- p.Probe(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	err := <-errCh
	assert.NotNil(t, err)
	assert.Contains(t, "context canceled", err.Error())
}

func TestProbe_WithCommand_WhenSuccessful(t *testing.T) {
	p := New(&config.Probe{TimeoutSeconds: 1, Exec: &config.ProbeExec{Command: []string{"true"}}})
	err := p.Probe(context.Background())
	assert.Nil(t, err)
}

func TestProbe_WithCommand_WhenFailed(t *testing.T) {
	p := New(&config.Probe{TimeoutSeconds: 1, Exec: &config.ProbeExec{Command: []string{"false"}}})
	err := p.Probe(context.Background())
	assert.NotNil(t, err)
}

func TestProbe_WithCommand_WhenTimedOut(t *testing.T) {
	p := New(&config.Probe{TimeoutSeconds: 1, Exec: &config.ProbeExec{Command: []string{"sleep", "2"}}})
	err := p.Probe(context.Background())
	assert.NotNil(t, err)
	assert.Contains(t, "context deadline exceeded", err.Error())
}

func TestRun_WhenCancelled(t *testing.T) {
	p := newWithProbeFunc(&config.Probe{
		TimeoutSeconds:   1,
		PeriodSeconds:    1,
		FailureThreshold: 1,
	}, fakeSucceedImmediately)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		errCh <- p.Run(ctx)
	}()

	time.Sleep(3 * time.Millisecond)
	cancel()

	err := <-errCh
	assert.NotNil(t, err)
	assert.Contains(t, "context canceled", err.Error())
}

func TestRun_WhenFailsEventually(t *testing.T) {
	sequence := []probeFunc{
		fakeSucceedImmediately,
		fakeSucceedImmediately,
		fakeFailImmediately,
	}
	ctx, probeFunc := withSequence(&sequence)
	p := newWithProbeFunc(&config.Probe{
		TimeoutSeconds:   1,
		PeriodSeconds:    1,
		FailureThreshold: 1,
	}, probeFunc)

	before := time.Now()
	err := p.Run(ctx)
	after := time.Now()
	diff := after.Sub(before).Truncate(time.Second)

	assert.NotNil(t, err)
	assert.Equal(t, errFakeFailed, err)
	assert.Equal(t, 2*time.Second, diff)
}

func TestRun_WithInitialDelay(t *testing.T) {
	p := newWithProbeFunc(&config.Probe{
		TimeoutSeconds:      1,
		PeriodSeconds:       1,
		FailureThreshold:    1,
		InitialDelaySeconds: 2,
	}, fakeFailImmediately)

	before := time.Now()
	err := p.Run(context.Background())
	after := time.Now()
	diff := after.Sub(before).Truncate(time.Second)

	assert.NotNil(t, err)
	assert.Equal(t, errFakeFailed, err)
	assert.Equal(t, 2*time.Second, diff)
}

func TestRun_WithFailureThreshold(t *testing.T) {
	sequence := []probeFunc{
		fakeFailImmediately,
		fakeSucceedImmediately,
		fakeFailImmediately,
		fakeFailImmediately,
	}
	ctx, probeFunc := withSequence(&sequence)
	p := newWithProbeFunc(&config.Probe{
		TimeoutSeconds:   1,
		PeriodSeconds:    1,
		FailureThreshold: 2,
	}, probeFunc)

	before := time.Now()
	err := p.Run(ctx)
	after := time.Now()
	diff := after.Sub(before).Truncate(time.Second)

	assert.NotNil(t, err)
	assert.Equal(t, errFakeFailed, err)
	assert.Equal(t, 3*time.Second, diff)
}

func TestRun_WithStartupThreshold_SurvivesStartup(t *testing.T) {
	sequence := []probeFunc{
		fakeFailImmediately,
		fakeFailImmediately,
		fakeSucceedImmediately,
		fakeFailImmediately,
	}
	ctx, probeFunc := withSequence(&sequence)
	p := newWithProbeFunc(&config.Probe{
		TimeoutSeconds:   1,
		PeriodSeconds:    1,
		FailureThreshold: 1,
		StartupThreshold: 3,
	}, probeFunc)

	before := time.Now()
	err := p.Run(ctx)
	after := time.Now()
	diff := after.Sub(before).Truncate(time.Second)

	assert.NotNil(t, err)
	assert.Equal(t, errFakeFailed, err)
	assert.Equal(t, 3*time.Second, diff)
}

func TestRun_WithStartupThreshold_FailsDuringStartup(t *testing.T) {
	sequence := []probeFunc{
		fakeFailImmediately,
		fakeFailImmediately,
		fakeFailImmediately,
		fakeSucceedImmediately,
	}
	ctx, probeFunc := withSequence(&sequence)
	p := newWithProbeFunc(&config.Probe{
		TimeoutSeconds:   1,
		PeriodSeconds:    1,
		FailureThreshold: 1,
		StartupThreshold: 3,
	}, probeFunc)

	before := time.Now()
	err := p.Run(ctx)
	after := time.Now()
	diff := after.Sub(before).Truncate(time.Second)

	assert.NotNil(t, err)
	assert.Equal(t, errFakeFailed, err)
	assert.Equal(t, 2*time.Second, diff)
}

func TestRun_WithTimeout(t *testing.T) {
	p := newWithProbeFunc(&config.Probe{
		TimeoutSeconds:   1,
		PeriodSeconds:    1,
		FailureThreshold: 1,
	}, withDelay(10*time.Second, fakeSucceedImmediately))

	err := p.Run(context.Background())
	assert.NotNil(t, err)
	assert.Contains(t, "context deadline exceeded", err.Error())
}
