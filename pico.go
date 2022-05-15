//go:build pico
// +build pico

package main

import (
	"machine"
)

func init() {
	spi0     = machine.SPI0
	sckPin   = machine.GPIO2
	sdoPin   = machine.GPIO3
	sdiPin   = machine.GPIO4
	csPin    = machine.GPIO5
	spi1     = machine.SPI1
	sck1Pin  = machine.GPIO10
	sdo1Pin  = machine.GPIO11
	sdi1Pin  = machine.GPIO12
	cs1Pin   = machine.GPIO13
	xrstPin  = machine.GPIO14
	xdcsPin  = machine.GPIO15
	xdreqPin = machine.GPIO16
	ledPin   = machine.LED
}