//go:build pico
// +build pico

package main

import (
    "machine"
)

func init() {
    spi      = machine.SPI0
    sckPin   = machine.GPIO2
    sdoPin   = machine.GPIO3
    sdiPin   = machine.GPIO4
    csSdPin  = machine.GPIO5
    csVsPin  = machine.GPIO6
    xrstPin  = machine.GPIO7
    xdcsPin  = machine.GPIO17
    xdreqPin = machine.GPIO16
    ledPin   = machine.LED
}
