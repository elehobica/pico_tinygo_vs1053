//go:build pico
// +build pico

package main

import (
	"machine"
)

func init() {
	spi      = machine.SPI1
	sckPin   = machine.GPIO10
	sdoPin   = machine.GPIO11
	sdiPin   = machine.GPIO12
	csSdPin  = machine.GPIO9
	csVsPin  = machine.GPIO13
	xrstPin  = machine.GPIO14
	xdcsPin  = machine.GPIO15
	xdreqPin = machine.GPIO16
	ledPin   = machine.LED
}