package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/radhus/voff/pkg/config"
	"github.com/radhus/voff/pkg/prober"
	"github.com/radhus/voff/pkg/watchdog"
	"github.com/radhus/voff/pkg/watchdog/dummy"
)

func failWithUsage(msg string) {
	fmt.Println("Error:", msg)
	fmt.Println()
	fmt.Println(
		"Usage:",
		os.Args[0],
		"-config /etc/voff.yaml",
	)
	flag.PrintDefaults()
	os.Exit(1)
}

func loop(ctx context.Context, device watchdog.Device, errCh <-chan error, interval time.Duration) {
	for {
		select {
		case <-ctx.Done():
			log.Print("Error: context closed: ", ctx.Err())
			return
		case err := <-errCh:
			log.Print("Error: probe failed: ", err)
			return
		case <-time.After(interval):
			err := device.Kick()
			if err != nil {
				log.Print("Error kicking watchdog, ignoring: ", err)
			}
			continue
		}
	}
}

func main() {
	configFile := flag.String("config", "", "Path to config file")
	dryRun := flag.Bool("dry-run", false, "Don't touch the watchdog device")
	flag.Parse()

	if *configFile == "" {
		failWithUsage("-config is mandatory")
	}

	configData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal("Couldn't read config file:", err)
	}
	config, err := config.ReadConfig(configData)
	if err != nil {
		log.Fatal("Couldn't parse config file:", err)
	}

	var device watchdog.Device
	if !*dryRun {
		devStat, err := os.Stat(config.Watchdog.Device)
		if err != nil {
			log.Fatal("Couldn't stat watchdog device:", err)
		}
		if devStat.Mode()&os.ModeDevice == 0 {
			log.Fatal("Watchdog device path must point to a device file!")
		}

		device, err = watchdog.Open(config.Watchdog.Device)
		if err != nil {
			log.Fatal("Couldn't open watchdog:", err)
		}
	} else {
		device = dummy.New("dummy")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error)

	for _, probe := range config.Probes {
		prober := prober.New(probe)
		go func() {
			errCh <- prober.Run(ctx)
		}()
	}

	interval := time.Duration(config.Watchdog.IntervalSeconds) * time.Second
	loop(ctx, device, errCh, interval)

	cancel()
	log.Fatal("Exiting")
}
