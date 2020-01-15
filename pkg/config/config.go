package config

import (
	"errors"

	yaml "gopkg.in/yaml.v1"
)

const (
	defaultFailureThreshold = 3
	defaultTimeoutSeconds   = 1
	defaultPeriodSeconds    = 10
	defaultSuccessThreshold = 1
	defaultWatchdogPath     = "/dev/watchdog"
	defaultIntervalSeconds  = 1
)

func setDefaultInt(current, fallback int) int {
	if current != 0 {
		return current
	}
	return fallback
}

type Watchdog struct {
	Device          string `yaml:"device"`
	IntervalSeconds int    `yaml:"intervalSeconds"`
}

func (w *Watchdog) populateDefaults() {
	if w.Device == "" {
		w.Device = defaultWatchdogPath
	}
	w.IntervalSeconds = setDefaultInt(w.IntervalSeconds, defaultIntervalSeconds)
}

func (w *Watchdog) validate() error {
	if len(w.Device) < 1 {
		return errors.New("Need to specify path to watchdog device")
	}
	if w.IntervalSeconds < 1 {
		return errors.New("Watchdog interval need to be at least 1")
	}
	return nil
}

type ProbeExec struct {
	Command []string `yaml:"command"`
}

// Probe mimicks the Probe type in k8s core/v1.
type Probe struct {
	// Actions
	Exec *ProbeExec `yaml:"exec"`
	// HTTPGet   interface{} `yaml:"httpGet"`
	// TCPSocket interface{} `yaml:"tcpSocket`

	FailureThreshold    int `yaml:"failureThreshold"`
	InitialDelaySeconds int `yaml:"initialDelaySeconds"`
	PeriodSeconds       int `yaml:"periodSeconds"`
	TimeoutSeconds      int `yaml:"timeoutSeconds"`
}

func (p *Probe) populateDefaults() {
	p.FailureThreshold = setDefaultInt(p.FailureThreshold, defaultFailureThreshold)
	p.PeriodSeconds = setDefaultInt(p.PeriodSeconds, defaultPeriodSeconds)
	p.TimeoutSeconds = setDefaultInt(p.TimeoutSeconds, defaultTimeoutSeconds)
}

func (p *Probe) validate() error {
	if p.Exec == nil || len(p.Exec.Command) < 1 {
		return errors.New("Probe need exec with at least one command")
	}
	if p.FailureThreshold < 1 {
		return errors.New("Probe failureThreshold needs to be at least 1")
	}
	if p.InitialDelaySeconds < 0 {
		return errors.New("Probe initialDelaySeconds need to be positive")
	}
	if p.PeriodSeconds < 1 {
		return errors.New("Probe periodSeconds need to be at least 1")
	}
	if p.TimeoutSeconds < 1 {
		return errors.New("Probe timeoutSeconds need to be at least 1")
	}
	return nil
}

type Config struct {
	Watchdog *Watchdog `yaml:"watchdog"`
	Probes   []*Probe  `yaml:"probes"`
}

func (c *Config) populateDefaults() {
	if c.Watchdog == nil {
		c.Watchdog = &Watchdog{}
	}
	c.Watchdog.populateDefaults()

	for _, probe := range c.Probes {
		probe.populateDefaults()
	}
}

func (c *Config) validate() error {
	if err := c.Watchdog.validate(); err != nil {
		return err
	}

	if len(c.Probes) < 1 {
		return errors.New("Need at least one probe")
	}
	for _, probe := range c.Probes {
		if err := probe.validate(); err != nil {
			return err
		}
	}
	return nil
}

func ReadConfig(data []byte) (*Config, error) {
	config := &Config{}
	err := yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	config.populateDefaults()
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}
