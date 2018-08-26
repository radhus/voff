package dummy

import (
	"fmt"
	"log"
	"os"

	"github.com/radhus/voff/pkg/watchdog"
)

type dummy struct {
	logger *log.Logger
}

func New(name string) watchdog.Device {
	return &dummy{
		logger: log.New(
			os.Stdout,
			fmt.Sprintf("[wd %s] ", name),
			0,
		),
	}
}

func (d *dummy) Close() error {
	return nil
}

func (d *dummy) Poke() error {
	d.logger.Println("Poke")
	return nil
}
