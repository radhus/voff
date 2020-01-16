# voff

Voff is a tool that kicks a watchdog device as long as no probe fail.

## Usage

Create configuration file:

```yaml
watchdog:
  device: /dev/watchdog
  intervalSeconds: 1
probes:
  - exec:
      command:
        - true
  - exec:
      command:
        - false
    failureThreshold: 10
    periodSeconds: 1
```

Run voff:

```bash
$ voff -config /etc/voff.yaml
```

## Configuration

Voff is configured with a YAML file with two top-level keys:

### `watchdog`

- `device`: watchdog device path.
- `intervalSeconds`: interval seconds to kick the watchdog.

### `probes`

A list of one or many probes, inspired by [Kubernetes container probes](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.10/#probe-v1-core),
where each probe can have the following keys:

- `exec`
  - `command`: list of command arguments to run, at least one required.
- `failureThreshold`: number of failed probes that should result in a failure. Default 1.
- `initialDelaySeconds`: number of seconds to wait before running the first probe. Default 0.
- `periodSeconds`: number of seconds to wait between each probe.
- `timeoutSeconds`: number of seconds to wait for the probe to complete.
