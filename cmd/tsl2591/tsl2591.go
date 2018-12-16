/**
 * tsl2591 - A command for interacting with TSL2591 lux sensors.
 */

package main

import (
	"log"
	"time"

	"github.com/mstahl/tsl2591"
)

func main() {
	tsl, err := tsl2591.NewTSL2591(&tsl2591.Opts{
		Gain:   tsl2591.TSL2591_GAIN_HIGH,
		Timing: tsl2591.TSL2591_INTEGRATIONTIME_600MS,
	})
	if err != nil {
		panic(err)
	}

	// tsl.Enable()

	ticker := time.NewTicker(10 * time.Second)

	for {
		// luminosity := tsl.TotalLuminosity()
		luminosity := tsl.GetFullLuminosity()
		log.Printf("0x%08x\n", luminosity)

		<-ticker.C
	}
}
