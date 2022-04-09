package main

import (
	"fmt"
	"machine"
	"time"

	"mylocal.com/fatfs"
	"mylocal.com/console"
	"tinygo.org/x/drivers/sdcard"
)

var (
	spi    machine.SPI
	sckPin machine.Pin
	sdoPin machine.Pin
	sdiPin machine.Pin
	csPin  machine.Pin
	ledPin machine.Pin

	serial  = machine.Serial
)

func main() {
	println(); println()
	println("======================")
	println("== pico_tinyfs_test ==")
	println("======================")

	led := ledPin
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.High()

	sd := sdcard.New(spi, sckPin, sdoPin, sdiPin, csPin)
	err := sd.Configure()
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		error_blink(led, 1)
	}

	filesystem := fatfs.New(&sd)

	// Configure FATFS with sector size (must match value in ff.h - use 512)
	filesystem.Configure(&fatfs.Config{
		SectorSize: 512,
	})

	err = filesystem.Mount()
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		error_blink(led, 2)
	}
	fmt.Printf("mount ok\r\n");

	fs_type := filesystem.GetFsType()
	fmt.Printf("Type is %s\r\n", fs_type.String())

	fmt.Printf("Card size: %7.2f GB (GB = 1E9 bytes)\r\n\r\n", float32(filesystem.GetCardSize()) * 1e-9)

	console.RunFor(&sd, filesystem)
}

func error_blink(led machine.Pin, count int) {
	for i := 0; i < count; i++ {
		led.High()
		time.Sleep(250 * time.Millisecond)
		led.Low()
		time.Sleep(250 * time.Millisecond)
	}
	led.Low()
	time.Sleep(500 * time.Millisecond)
}