//go:build pico
// +build pico

package main

import (
	"machine"
	"github.com/elehobica/pico_tinygo_vs1053/mymachine"
)

func init() {
	spi = mymachine.SPI{machine.SPI0}
	sckPin = machine.GP2
	sdoPin = machine.GP3
	sdiPin = machine.GP4
	csPin = machine.GP5

	ledPin = machine.LED
}