package main

import (
	"fmt"
	"machine"
	"time"
	"os"

	//"tinygo.org/x/drivers/sdcard"
	"pico_tinygo_fatfs_test/sdcard"
	//"tinygo.org/x/tinyfs/fatfs"
	"pico_tinygo_fatfs_test/fatfs"
	"pico_tinygo_fatfs_test/mymachine"
)

var (
	spi    mymachine.SPI
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

func (pin Pin) ErrorBlinkFor(count int) {
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

func (pin Pin) OkBlinkFor() {
	for {
		pin.High()
		time.Sleep(1000 * time.Millisecond)
		pin.Low()
		time.Sleep(1000 * time.Millisecond)
	}
}

type TestError struct {
	error
	Code int
}

func main() {
	led := &Pin{&ledPin}
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.High()

	err := fatfs_test(led)
	if err != nil {
		fmt.Printf("ERROR[%d]: %s\r\n", err.Code, err.Error())
		led.ErrorBlinkFor(err.Code)
	}

	led.OkBlinkFor()
}

func fatfs_test(led *Pin) (testError *TestError) {
	// SPI BaudRate
	const SPI_BAUDRATE_MHZ = 50

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

	sd := sdcard.New(spi, sckPin, sdoPin, sdiPin, csPin)
	err := sd.Configure()
	if err != nil {
		return &TestError{ error: fmt.Errorf("configure error: %s", err.Error()), Code: 1 }
	}

	// Set SPI clock speed (not effective if set before here)
	spi.SetBaudRate(SPI_BAUDRATE_MHZ * machine.MHz)

	filesystem := fatfs.New(&sd)

	// Configure FATFS with sector size (must match value in ff.h - use 512)
	filesystem.Configure(&fatfs.Config{
		SectorSize: 512,
	})

	err = filesystem.Mount()
	if err != nil {
		return &TestError{ error: fmt.Errorf("mount error: %s", err.Error()), Code: 2 }
	}
	fmt.Printf("mount ok\r\n")

	fs_type, err := filesystem.GetFsType()
	fmt.Printf("Type is %s\r\n", fs_type.String())

	size, err := filesystem.GetCardSize()
	fmt.Printf("Card size: %7.2f GB (GB = 1E9 bytes)\r\n\r\n", float32(size) * 1e-9)

	f, err := filesystem.OpenFile("bench.dat", os.O_RDWR | os.O_CREATE | os.O_TRUNC)
	if err != nil {
		return &TestError{ error: fmt.Errorf("open error: %s", err.Error()), Code: 3 }
	}
	defer f.Close()

	// get *fatfs.File type from tinyfs.File interface (Type Assertion)
	ff, ok := f.(*fatfs.File)
	if ok != true {
		return &TestError{ error: fmt.Errorf("conversion to *fatfs.File failed"), Code: 4 }
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
	n := int64(FILE_SIZE / BUF_SIZE)

	//----------------
	// do write test
	//----------------
	fmt.Printf("Starting write test, please wait.\r\n\r\n")
	fmt.Printf("write speed and latency\r\n")
	fmt.Printf("speed,max,min,avg\r\n")
	fmt.Printf("KB/Sec,usec,usec,usec\r\n")

	for nTest := 0; nTest < WRITE_COUNT; nTest++ {
		err = ff.Seek(0)
		if err != nil {
			return &TestError{ error: fmt.Errorf("seek error: %s", err.Error()), Code: 5 }
		}
		err = ff.Truncate()
		if err != nil {
			return &TestError{ error: fmt.Errorf("truncate error: %s", err.Error()), Code: 6 }
		}
		if PRE_ALLOCATE {
			err = ff.Expand(FILE_SIZE, false)
			if err != nil {
				return &TestError{ error: fmt.Errorf("preallocate error: %s", err.Error()), Code: 7 }
			}
		}
		maxLatency := int64(0)
		minLatency := int64(9999999)
		totalLatency := int64(0)
		skipLatency := SKIP_FIRST_LATENCY
		t := time.Since(start).Milliseconds()
		for i := int64(0); i < n; i++ {
			m := time.Since(start).Microseconds()
			bw, err := ff.Write(buf)
			if err != nil || bw != BUF_SIZE {
				return &TestError{ error: fmt.Errorf("write failed: %s %d", err.Error(), bw), Code: 8 }
			}
			m = time.Since(start).Microseconds() - m
			totalLatency += m
			if skipLatency {
				// Wait until first write to SD, not just a copy to the cache.
				pos, err := ff.Tell()
				if err != nil {
					return &TestError{ error: fmt.Errorf("tell error: %s", err.Error()), Code: 9 }
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
			return &TestError{ error: fmt.Errorf("sync failed: %s", err.Error()), Code: 10 }
		}
		t = time.Since(start).Milliseconds() - t
		s, err := ff.Size()
		if err != nil {
			return &TestError{ error: fmt.Errorf("size error: %s", err.Error()), Code: 11 }
		}
		fmt.Printf("%7.4f, %d, %d, %d\r\n", float32(s)/float32(t), maxLatency, minLatency, totalLatency/n)
	}
	fmt.Printf("\r\n")

	//----------------
	// do read test
	//----------------
	fmt.Printf("Starting read test, please wait.\r\n\r\n")
	fmt.Printf("read speed and latency\r\n")
	fmt.Printf("speed,max,min,avg\r\n")
	fmt.Printf("KB/Sec,usec,usec,usec\r\n")

	for nTest := 0; nTest < WRITE_COUNT; nTest++ {
		err = ff.Rewind()
		if err != nil {
			return &TestError{ error: fmt.Errorf("rewind failed: %s", err.Error()), Code: 12 }
		}
		maxLatency := int64(0)
		minLatency := int64(9999999)
		totalLatency := int64(0)
		skipLatency := SKIP_FIRST_LATENCY
		t := time.Since(start).Milliseconds()
		for i := int64(0); i < n; i++ {
			buf[BUF_SIZE - 1] = 0
			m := time.Since(start).Microseconds()
			br, err := ff.Read(buf)
			if err != nil || br != BUF_SIZE {
				return &TestError{ error: fmt.Errorf("read failed: %s %d", err.Error(), br), Code: 13 }
			}
			m = time.Since(start).Microseconds() - m
			totalLatency += m
			if buf[BUF_SIZE - 1] != '\n' {
				return &TestError{ error: fmt.Errorf("data check error"), Code: 14 }
			}
			if skipLatency {
				skipLatency = false
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
		t = time.Since(start).Milliseconds() - t
		s, err := ff.Size()
		if err != nil {
			return &TestError{ error: fmt.Errorf("size error: %s", err.Error()), Code: 15 }
		}
		fmt.Printf("%7.4f, %d, %d, %d\r\n", float32(s)/float32(t), maxLatency, minLatency, totalLatency/n)
	}
	fmt.Printf("\r\nDone\r\n")

	return nil
}