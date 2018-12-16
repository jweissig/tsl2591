/**
 * tsl2591 - A command for interacting with TSL2591 lux sensors.
 */

package main

import (
	"log"
	"time"

	"github.com/mstahl/tsl2591"
)

const Interval = 1 * time.Second

func main() {
	tsl, err := tsl2591.NewTSL2591(&tsl2591.Opts{
		Gain:   tsl2591.TSL2591_GAIN_LOW,
		Timing: tsl2591.TSL2591_INTEGRATIONTIME_600MS,
	})
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(Interval)

	for {
		channel0, channel1 := tsl.GetFullLuminosity()
		log.Printf("0x%04x 0x%04x\n", channel0, channel1)
		<-ticker.C
	}
}
