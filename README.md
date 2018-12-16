This is a Golang driver for the TSL2591 lux sensor.

## Installation

    go get -u github.com/mstahl/tsl2591

## Usage

    import "github.com/mstahl/tsl2591"

For now, `tsl2591` only supports retrieving luminosity data, so no interrupts
or alerts yet.

```go
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
```

## Acknowledgements

This library is basically a golang port of [Adafruit's TSL2591 library](https://github.com/adafruit/Adafruit_TSL2591_Library/)
