package prober

import (
	"context"
	"os/exec"
	"time"

	"github.com/radhus/voff/pkg/config"
)

type Prober interface {
	Probe(context.Context) error
	Run(context.Context) error
}

type probeFunc func(context.Context) error

type prober struct {
	config *config.Probe
	probe  probeFunc
}

func New(probeConfig *config.Probe) Prober {
	p := &prober{config: probeConfig}
	p.probe = p.exec
	return p
}

func (p *prober) exec(ctx context.Context) error {
	name := p.config.Exec.Command[0]
	args := p.config.Exec.Command[1:]
	return exec.CommandContext(ctx, name, args...).Run()
}

func (p *prober) Probe(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.config.TimeoutSeconds)*time.Second)
	defer cancel()

	errCh := make(chan error)
	go func() {
		errCh <- p.probe(ctx)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (p *prober) Run(ctx context.Context) error {
	if p.config.InitialDelaySeconds > 0 {
		time.Sleep(time.Duration(p.config.InitialDelaySeconds) * time.Second)
	}

	periodDuration := time.Duration(p.config.PeriodSeconds) * time.Second
	failureCount := 0
	for {
		err := p.Probe(ctx)
		if err != nil {
			failureCount++
		} else {
			failureCount = 0
		}

		if failureCount >= p.config.FailureThreshold {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(periodDuration):
			continue
		}
	}
}
