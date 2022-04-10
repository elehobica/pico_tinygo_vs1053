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

type Pin struct {
	*machine.Pin
}

func (pin Pin) Toggle() {
	pin.Set(!pin.Get())
}

func (pin Pin) ErrorBlink(count int) {
	for {
		for i := 0; i < count; i++ {
			pin.High()
			time.Sleep(250 * time.Millisecond)
			pin.Low()
			time.Sleep(250 * time.Millisecond)
		}
		pin.Low()
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	// Set PRE_ALLOCATE true to pre-allocate file clusters.
	const PRE_ALLOCATE = true

	// Set SKIP_FIRST_LATENCY true if the first read/write to the SD can
	// be avoid by writing a file header or reading the first record.
	const SKIP_FIRST_LATENCY = true

	// Size of read/write.
	const BUF_SIZE = 512

	// File size in MB where MB = 1,000,000 bytes.
	const FILE_SIZE_MB = 5

	// Write pass count.
	const WRITE_COUNT = 2

	// File size in bytes.
	const FILE_SIZE = 1000000*FILE_SIZE_MB

	var buf []byte

	start := time.Now()

	println(); println()
	println("============================")
	println("== pico_tinygo_fatfs_test ==")
	println("============================")

	led := &Pin{&ledPin}
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.High()

	sd := sdcard.New(spi, sckPin, sdoPin, sdiPin, csPin)
	err := sd.Configure()
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		led.ErrorBlink(1)
	}

	filesystem := fatfs.New(&sd)

	// Configure FATFS with sector size (must match value in ff.h - use 512)
	filesystem.Configure(&fatfs.Config{
		SectorSize: 512,
	})

	err = filesystem.Mount()
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		led.ErrorBlink(2)
	}
	fmt.Printf("mount ok\r\n")

	fs_type := filesystem.GetFsType()
	fmt.Printf("Type is %s\r\n", fs_type.String())

	fmt.Printf("Card size: %7.2f GB (GB = 1E9 bytes)\r\n\r\n", float32(filesystem.GetCardSize()) * 1e-9)

	f, err := filesystem.OpenFile("bench.dat", os.O_RDWR | os.O_CREATE | os.O_TRUNC)
	if err != nil {
		fmt.Printf("open error %s\r\n", err.Error())
		led.ErrorBlink(3)
	}
	defer f.Close()

	// get *fatfs.File type from tinyfs.File interface (Type Assertion)
	ff, ok := f.(*fatfs.File)
	if ok != true {
		fmt.Printf("conversion to *fatfs.File failed\r\n")
		led.ErrorBlink(4)
	}

	// fill buf with known data
	if BUF_SIZE > 1 {
		for i := 0; i < (BUF_SIZE + 3) / 4 * 4 - 2; i++ {
			buf = append(buf, 'A' + uint8(i % 26))
		}
		buf = append(buf, '\r')
	}
	buf = append(buf, '\n')

	fmt.Printf("FILE_SIZE_MB = %d\r\n", FILE_SIZE_MB)
    fmt.Printf("BUF_SIZE = %d bytes\r\n", BUF_SIZE)
    fmt.Printf("Starting write test, please wait.\r\n\r\n")

	//----------------
	// do write test
	//----------------
	fmt.Printf("write speed and latency\r\n")
	fmt.Printf("speed,max,min,avg\r\n")
	fmt.Printf("KB/Sec,usec,usec,usec\r\n")

	for nTest := 0; nTest < WRITE_COUNT; nTest++ {
		err = ff.Seek(0)
		if err != nil {
			fmt.Printf("seek error %s\r\n", err.Error())
			led.ErrorBlink(5)
		}
		err = ff.Truncate()
		if err != nil {
			fmt.Printf("truncate error %s\r\n", err.Error())
			led.ErrorBlink(6)
		}
		if PRE_ALLOCATE {
			err = ff.Expand(FILE_SIZE, false)
			if err != nil {
				fmt.Printf("preallocate error %s\r\n", err.Error())
				led.ErrorBlink(7)
			}
		}
		maxLatency := int64(0)
        minLatency := int64(9999999)
        totalLatency := int64(0)
        skipLatency := SKIP_FIRST_LATENCY
		n := int64(FILE_SIZE / BUF_SIZE)
		t := time.Since(start).Milliseconds()
		for i := int64(0); i < n; i++ {
			m := time.Since(start).Microseconds()
			bw, err := ff.Write(buf)
			if err != nil || bw != BUF_SIZE {
				fmt.Printf("write failed %s %d\r\n", err.Error(), bw)
				led.ErrorBlink(8)
			}
			m = time.Since(start).Microseconds() - m
			totalLatency += m
			if skipLatency {
                // Wait until first write to SD, not just a copy to the cache.
				pos, err := ff.Tell()
				if err != nil {
					fmt.Printf("tell error %s\r\n", err.Error())
					led.ErrorBlink(9)
				}
                skipLatency = pos < 512
            } else {
                if maxLatency < m {
                    maxLatency = m
                }
                if minLatency > m {
                    minLatency = m
                }
            }
			if i % 10 == 0 {
				led.Toggle()
			}
		}
		err = ff.Sync()
		if err != nil {
			fmt.Printf("sync failed %s\r\n", err.Error())
			led.ErrorBlink(10)
		}
		t = time.Since(start).Milliseconds() - t
		s, err := ff.Size()
		if err != nil {
			fmt.Printf("size error %s\r\n", err.Error())
			led.ErrorBlink(11)
		}
		fmt.Printf("%7.4f, %d, %d, %d\r\n", float32(s)/float32(t), maxLatency, minLatency, totalLatency/n)
	}

	console.RunFor(&sd, filesystem)
}