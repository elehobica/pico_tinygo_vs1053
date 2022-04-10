package main

import (
	"fmt"
	"machine"
	"time"
	"os"

	"github.com/elehobica/pico_tinygo_fatfs_test/fatfs"
	"github.com/elehobica/pico_tinygo_fatfs_test/console"
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
	// Set PRE_ALLOCATE true to pre-allocate file clusters.
	const PRE_ALLOCATE = true;

	// Size of read/write.
	const BUF_SIZE = 512

	// File size in MB where MB = 1,000,000 bytes.
	const FILE_SIZE_MB = 5;

	// Write pass count.
	const WRITE_COUNT = 2;

	// File size in bytes.
	const FILE_SIZE = 1000000*FILE_SIZE_MB;

	var buf[(BUF_SIZE + 3) / 4 * 4] uint8

	println(); println()
	println("============================")
	println("== pico_tinygo_fatfs_test ==")
	println("============================")

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

	f, err := filesystem.OpenFile("bench.dat", os.O_RDWR | os.O_CREATE | os.O_TRUNC)
	if err != nil {
		fmt.Printf("open error %s\r\n", err.Error())
		error_blink(led, 3)
	}
	defer f.Close()

	// get *fatfs.File type from tinyfs.File interface (Type Assertion)
	ff, ok := f.(*fatfs.File)
	if ok != true {
		fmt.Printf("conversion to *fatfs.File failed\r\n")
		error_blink(led, 4)
	}

	// fill buf with known data
	if BUF_SIZE > 1 {
		for i := 0; i < BUF_SIZE - 2; i++ {
			buf[i] = 'A' + uint8(i % 26)
		}
		buf[BUF_SIZE - 2] = '\r'
	}
	buf[BUF_SIZE - 1] = '\n'

	fmt.Printf("FILE_SIZE_MB = %d\r\n", FILE_SIZE_MB);
    fmt.Printf("BUF_SIZE = %d bytes\r\n", BUF_SIZE);
    fmt.Printf("Starting write test, please wait.\r\n\r\n");

	// do write test
	//n := FILE_SIZE / BUF_SIZE;
	fmt.Printf("write speed and latency\r\n");
	fmt.Printf("speed,max,min,avg\r\n");
	fmt.Printf("KB/Sec,usec,usec,usec\r\n");

	for nTest := 0; nTest < WRITE_COUNT; nTest++ {
		err = ff.Seek(0)
		if err != nil {
			fmt.Printf("seek error %s\r\n", err.Error())
			error_blink(led, 5)
		}
		err = ff.Truncate()
		if err != nil {
			fmt.Printf("truncate error %s\r\n", err.Error())
			error_blink(led, 6)
		}
		if PRE_ALLOCATE {
			err = ff.Expand(FILE_SIZE, false)
			if err != nil {
				fmt.Printf("preallocate error %s\r\n", err.Error())
				error_blink(led, 7)
			}
		}
	}

	console.RunFor(&sd, filesystem)
}

func error_blink(led machine.Pin, count int) {
	for {
		for i := 0; i < count; i++ {
			led.High()
			time.Sleep(250 * time.Millisecond)
			led.Low()
			time.Sleep(250 * time.Millisecond)
		}
		led.Low()
		time.Sleep(500 * time.Millisecond)
	}
}