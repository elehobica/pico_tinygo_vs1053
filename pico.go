//go:build pico
// +build pico

package main

import (
	"machine"
)

func init() {
	spi = *machine.SPI0
	sckPin = machine.GP2
	sdoPin = machine.GP3
	sdiPin = machine.GP4
	csPin = machine.GP5

	ledPin = machine.LED
}