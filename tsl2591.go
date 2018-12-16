/**
 * tsl2591 - Package for interacting with TSL2591 lux sensors.
 *
 * Mostly ported from Adafruit's library here:
 * https://github.com/adafruit/Adafruit_TSL2591_Library
 */
package tsl2591

import (
	"errors"
	"log"
	"time"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
)

// General purpose consts
const (
	TSL2591_VISIBLE      byte = 2 ///< channel 0 - channel 1
	TSL2591_INFRARED     byte = 1 ///< channel 1
	TSL2591_FULLSPECTRUM byte = 0 ///< channel 0

	TSL2591_ADDR uint16 = 0x29 ///< Default I2C address

	TSL2591_COMMAND_BIT byte = 0xA0 ///< 1010 0000: bits 7 and 5 for 'command normal'

	///! Special Function Command for "Clear ALS and no persist ALS interrupt"
	TSL2591_CLEAR_INT byte = 0xE7
	///! Special Function Command for "Interrupt set - forces an interrupt"
	TSL2591_TEST_INT byte = 0xE4

	TSL2591_WORD_BIT  byte = 0x20 ///< 1 = read/write word rather than byte
	TSL2591_BLOCK_BIT byte = 0x10 ///< 1 = using block read/write

	TSL2591_ENABLE_POWEROFF byte = 0x00 ///< Flag for ENABLE register to disable
	TSL2591_ENABLE_POWERON  byte = 0x01 ///< Flag for ENABLE register to enable
	TSL2591_ENABLE_AEN      byte = 0x02 ///< ALS Enable. This field activates ALS function. Writing a one activates the ALS. Writing a zero disables the ALS.
	TSL2591_ENABLE_AIEN     byte = 0x10 ///< ALS Interrupt Enable. When asserted permits ALS interrupts to be generated, subject to the persist filter.
	TSL2591_ENABLE_NPIEN    byte = 0x80 ///< No Persist Interrupt Enable. When asserted NP Threshold conditions will generate an interrupt, bypassing the persist filter

	TSL2591_LUX_DF    float64 = 408.0 ///< Lux cooefficient
	TSL2591_LUX_COEFB float64 = 1.64  ///< CH0 coefficient
	TSL2591_LUX_COEFC float64 = 0.59  ///< CH1 coefficient A
	TSL2591_LUX_COEFD float64 = 0.86  ///< CH2 coefficient B
)

// TSL2591 Register map
const (
	TSL2591_REGISTER_ENABLE            byte = 0x00 // Enable register
	TSL2591_REGISTER_CONTROL           byte = 0x01 // Control register
	TSL2591_REGISTER_THRESHOLD_AILTL   byte = 0x04 // ALS low threshold lower byte
	TSL2591_REGISTER_THRESHOLD_AILTH   byte = 0x05 // ALS low threshold upper byte
	TSL2591_REGISTER_THRESHOLD_AIHTL   byte = 0x06 // ALS high threshold lower byte
	TSL2591_REGISTER_THRESHOLD_AIHTH   byte = 0x07 // ALS high threshold upper byte
	TSL2591_REGISTER_THRESHOLD_NPAILTL byte = 0x08 // No Persist ALS low threshold lower byte
	TSL2591_REGISTER_THRESHOLD_NPAILTH byte = 0x09 // No Persist ALS low threshold higher byte
	TSL2591_REGISTER_THRESHOLD_NPAIHTL byte = 0x0A // No Persist ALS high threshold lower byte
	TSL2591_REGISTER_THRESHOLD_NPAIHTH byte = 0x0B // No Persist ALS high threshold higher byte
	TSL2591_REGISTER_PERSIST_FILTER    byte = 0x0C // Interrupt persistence filter
	TSL2591_REGISTER_PACKAGE_PID       byte = 0x11 // Package Identification
	TSL2591_REGISTER_DEVICE_ID         byte = 0x12 // Device Identification
	TSL2591_REGISTER_DEVICE_STATUS     byte = 0x13 // Internal Status
	TSL2591_REGISTER_CHAN0_LOW         byte = 0x14 // Channel 0 data, low byte
	TSL2591_REGISTER_CHAN0_HIGH        byte = 0x15 // Channel 0 data, high byte
	TSL2591_REGISTER_CHAN1_LOW         byte = 0x16 // Channel 1 data, low byte
	TSL2591_REGISTER_CHAN1_HIGH        byte = 0x17 // Channel 1 data, high byte
)

// Constants for adjusting the sensor integration timing
const (
	TSL2591_INTEGRATIONTIME_100MS byte = 0x00 // 100 millis
	TSL2591_INTEGRATIONTIME_200MS byte = 0x01 // 200 millis
	TSL2591_INTEGRATIONTIME_300MS byte = 0x02 // 300 millis
	TSL2591_INTEGRATIONTIME_400MS byte = 0x03 // 400 millis
	TSL2591_INTEGRATIONTIME_500MS byte = 0x04 // 500 millis
	TSL2591_INTEGRATIONTIME_600MS byte = 0x05 // 600 millis
)

// Constants for adjusting the persistance filter (for interrupts)
const (
	TSL2591_PERSIST_EVERY byte = 0x00 // Every ALS cycle generates an interrupt
	TSL2591_PERSIST_ANY   byte = 0x01 // Any value outside of threshold range
	TSL2591_PERSIST_2     byte = 0x02 // 2 consecutive values out of range
	TSL2591_PERSIST_3     byte = 0x03 // 3 consecutive values out of range
	TSL2591_PERSIST_5     byte = 0x04 // 5 consecutive values out of range
	TSL2591_PERSIST_10    byte = 0x05 // 10 consecutive values out of range
	TSL2591_PERSIST_15    byte = 0x06 // 15 consecutive values out of range
	TSL2591_PERSIST_20    byte = 0x07 // 20 consecutive values out of range
	TSL2591_PERSIST_25    byte = 0x08 // 25 consecutive values out of range
	TSL2591_PERSIST_30    byte = 0x09 // 30 consecutive values out of range
	TSL2591_PERSIST_35    byte = 0x0A // 35 consecutive values out of range
	TSL2591_PERSIST_40    byte = 0x0B // 40 consecutive values out of range
	TSL2591_PERSIST_45    byte = 0x0C // 45 consecutive values out of range
	TSL2591_PERSIST_50    byte = 0x0D // 50 consecutive values out of range
	TSL2591_PERSIST_55    byte = 0x0E // 55 consecutive values out of range
	TSL2591_PERSIST_60    byte = 0x0F // 60 consecutive values out of range
)

// Constants for adjusting the sensor gain
const (
	TSL2591_GAIN_LOW  byte = 0x00 /// low gain (1x)
	TSL2591_GAIN_MED  byte = 0x10 /// medium gain (25x)
	TSL2591_GAIN_HIGH byte = 0x20 /// medium gain (428x)
	TSL2591_GAIN_MAX  byte = 0x30 /// max gain (9876x)
)

type Opts struct {
	Gain   byte
	Timing byte
}

type TSL2591 struct {
	enabled bool
	timing  byte
	gain    byte
	dev     *i2c.Dev
}

// Begin sets up a TSL2591 chip via the I2C protocol, sets its gain and timing
// attributes, and returns an error if any occurred in that process or if the
// TSL2591 was not found.
func NewTSL2591(opts *Opts) (*TSL2591, error) {
	bus, err := i2creg.Open("")
	if err != nil {
		panic(err)
	}

	tsl := &TSL2591{dev: &i2c.Dev{Bus: bus, Addr: TSL2591_ADDR}}

	// Read the device ID from the TSL2591. It should be 0x50
	id, err := tsl.read8(TSL2591_COMMAND_BIT | TSL2591_REGISTER_DEVICE_ID)
	if err != nil {
		panic(err)
	}
	if id != 0x50 {
		return nil, errors.New("Can't find a TSL2591 on I2C bus /dev/i2c-1")
	}

	tsl.SetTiming(opts.Timing)
	tsl.SetGain(opts.Gain)

	// TSL2591 boots up disabld
	tsl.Disable()

	return tsl, nil
}

func (tsl *TSL2591) Enable() {
	var write []byte = []byte{
		TSL2591_COMMAND_BIT | TSL2591_REGISTER_ENABLE,
		TSL2591_ENABLE_POWERON | TSL2591_ENABLE_AEN | TSL2591_ENABLE_AIEN | TSL2591_ENABLE_NPIEN,
	}
	if err := tsl.dev.Tx(write, nil); err != nil {
		panic(err)
	}
	tsl.enabled = true
}

func (tsl *TSL2591) Disable() {
	var write []byte = []byte{
		TSL2591_COMMAND_BIT | TSL2591_REGISTER_ENABLE,
		TSL2591_ENABLE_POWEROFF,
	}
	if err := tsl.dev.Tx(write, nil); err != nil {
		panic(err)
	}
	tsl.enabled = false
}

func (tsl *TSL2591) SetGain(gain byte) {
	tsl.Enable()

	write := []byte{
		TSL2591_COMMAND_BIT | TSL2591_REGISTER_CONTROL,
		tsl.timing | gain,
	}
	if err := tsl.dev.Tx(write, nil); err != nil {
		panic(err)
	}

	tsl.Disable()
	tsl.gain = gain
}

func (tsl *TSL2591) SetTiming(timing byte) {
	tsl.Enable()

	write := []byte{
		TSL2591_COMMAND_BIT | TSL2591_REGISTER_CONTROL,
		timing | tsl.gain,
	}
	if err := tsl.dev.Tx(write, nil); err != nil {
		panic(err)
	}

	tsl.Disable()
	tsl.timing = timing
}

// TODO
func CalculateLux(ch0, ch1 uint16) float64 {
	return 0.0
}

func (tsl *TSL2591) GetFullLuminosity() uint32 {
	tsl.Enable()

	// Delay for ADC to complete
	for d := byte(0); d < tsl.timing; d++ {
		time.Sleep(120 * time.Millisecond)
	}

	var x uint32
	for addr := TSL2591_REGISTER_CHAN0_LOW; addr <= TSL2591_REGISTER_CHAN1_HIGH; addr++ {
		b, err := tsl.read8(TSL2591_COMMAND_BIT | addr)
		if err != nil {
			panic(err)
		}
		x = x<<8 | uint32(b)
	}

	tsl.Disable()

	return x
}

// Low-level private methods

// TODO If there are weird readings coming out, it could be because of the
// return value here.
func (tsl *TSL2591) read16(cmd byte) (uint16, error) {
	// read := make([]byte, 2)
	var read [2]byte
	if err := tsl.dev.Tx([]byte{cmd}, read[:]); err != nil {
		return 0, err
	}
	return uint16(read[0]<<8) | uint16(read[1]), nil
}

func (tsl *TSL2591) read8(cmd byte) (byte, error) {
	// read := make([]byte, 1)
	var read [1]byte
	if err := tsl.dev.Tx([]byte{cmd}, read[:]); err != nil {
		return 0, err
	}
	log.Println("Received", read)
	return read[0], nil
}
