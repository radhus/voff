package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig_ValidFull(t *testing.T) {
	data := []byte(`
watchdog:
  device: /dev/wd
  intervalSeconds: 3
probes:
  - exec:
      command:
        - true
    failureThreshold: 5
    initialDelaySeconds: 3
    periodSeconds: 4
    timeoutSeconds: 6
  - exec:
      command:
        - true
    failureThreshold: 5
    initialDelaySeconds: 3
    periodSeconds: 4
    timeoutSeconds: 6
`)

	config, err := ReadConfig(data)
	assert.Nil(t, err)
	assert.NotNil(t, config)

	assert.NotNil(t, config.Watchdog)
	assert.Equal(t, "/dev/wd", config.Watchdog.Device)
	assert.Equal(t, 3, config.Watchdog.IntervalSeconds)

	assert.Len(t, config.Probes, 2)
	for _, probe := range config.Probes {
		assert.NotNil(t, probe.Exec)
		assert.Equal(t, []string{"true"}, probe.Exec.Command)
		assert.Equal(t, 5, probe.FailureThreshold)
		assert.Equal(t, 3, probe.InitialDelaySeconds)
		assert.Equal(t, 4, probe.PeriodSeconds)
		assert.Equal(t, 6, probe.TimeoutSeconds)
	}
}

func TestReadConfig_ValidDefaults(t *testing.T) {
	data := []byte(`
probes:
  - exec:
      command:
        - true
  - exec:
      command:
        - true
`)
	config, err := ReadConfig(data)
	assert.Nil(t, err)
	assert.NotNil(t, config)

	assert.NotNil(t, config.Watchdog)
	assert.Equal(t, "/dev/watchdog", config.Watchdog.Device)
	assert.Equal(t, 1, config.Watchdog.IntervalSeconds)

	assert.Len(t, config.Probes, 2)
	for _, probe := range config.Probes {
		assert.NotNil(t, probe.Exec)
		assert.Equal(t, []string{"true"}, probe.Exec.Command)
		assert.Equal(t, 3, probe.FailureThreshold)
		assert.Equal(t, 0, probe.InitialDelaySeconds)
		assert.Equal(t, 10, probe.PeriodSeconds)
		assert.Equal(t, 1, probe.TimeoutSeconds)
	}
}
