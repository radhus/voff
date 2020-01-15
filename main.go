package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/radhus/voff/pkg/watchdog"
	"github.com/radhus/voff/pkg/watchdog/dummy"
)

func triggerCheck(cmd string) bool {
	fork := exec.Command("sh", "-c", cmd)
	output, err := fork.CombinedOutput()
	log.Printf("Command output:\n%s\n", string(output))
	return err == nil
}

func failWithUsage(msg string) {
	fmt.Println("Error:", msg)
	fmt.Println()
	fmt.Println(
		"Usage:",
		os.Args[0],
		"-device /dev/watchdog -check \"ping -c 1 127.0.0.1\"",
	)
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	devicePath := flag.String("device", "/dev/watchdog", "Watchdog device")
	dryRun := flag.Bool("dry-run", false, "Don't touch the watchdog device")
	check := flag.String("check", "", "Command to execute to check status")
	interval := flag.Int("interval", 60, "Interval (seconds) to check and kick the watchdog")

	flag.Parse()

	if *check == "" {
		failWithUsage("-check is mandatory")
	}
	if !triggerCheck(*check) {
		log.Fatalln("Command failed initial check")
	}

	var device watchdog.Device
	if !*dryRun {
		devStat, err := os.Stat(*devicePath)
		if err != nil {
			log.Fatal("Couldn't stat watchdog device:", err)
		}
		if devStat.Mode()&os.ModeDevice == 0 {
			failWithUsage("-device must point to a device file")
		}

		device, err = watchdog.Open(*devicePath)
		if err != nil {
			log.Fatal("Couldn't open watchdog:", err)
		}
	} else {
		device = dummy.New("dummy")
	}

	tick := time.Tick(time.Duration(*interval) * time.Second)
	for range tick {
		log.Println("Running check...")
		if triggerCheck(*check) {
			log.Println("Check successful, poking watchdog")
			device.Kick()
		} else {
			log.Println("Check unsuccessful!")
		}
	}
}
