package main

import (
	"fmt"
	"machine"
	"time"
	"os"

	//"tinygo.org/x/drivers/sdcard"
	"github.com/elehobica/pico_tinygo_vs1053/sdcard"
	"tinygo.org/x/tinyfs"
	//"tinygo.org/x/tinyfs/fatfs"
	"github.com/elehobica/pico_tinygo_vs1053/fatfs"
	"github.com/elehobica/pico_tinygo_vs1053/mymachine"
	"github.com/elehobica/pico_tinygo_vs1053/vs1053"
)

var (
	spi      *machine.SPI
	sckPin   machine.Pin
	sdoPin   machine.Pin
	sdiPin   machine.Pin
	csSdPin  machine.Pin
	csVsPin  machine.Pin
	xrstPin  machine.Pin
	xdcsPin  machine.Pin
	xdreqPin machine.Pin
	ledPin   machine.Pin
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

	// Create SPI devices
	mySpi := mymachine.NewSPI(spi)
	mySpi.Configure(machine.SPIConfig{
		SCK:       sckPin,
		SDO:       sdoPin,
		SDI:       sdiPin,
		LSBFirst:  false,
		Mode:      0, // phase=0, polarity=0
	})
	codec := vs1053.New(mySpi, csVsPin, xrstPin, xdcsPin, xdreqPin)
	sd := sdcard.New(mySpi, csSdPin)

	// Start Test
	err := vs1053_test(led, codec, sd)
	if err != nil {
		fmt.Printf("ERROR[%d]: %s\r\n", err.Code, err.Error())
		led.ErrorBlinkFor(err.Code)
	}

	led.OkBlinkFor()
}

func vs1053_test(led *Pin, codec vs1053.Device, sd sdcard.Device) (testError *TestError) {
	println(); println()
	println("========================")
	println("== pico_tinygo_vs1053 ==")
	println("========================")

	err := codec.Configure()
	if err != nil {
		return &TestError{ error: fmt.Errorf("codec configure error: %s", err.Error()), Code: 1 }
	}
	codec.SwitchToMp3Mode()

	err = sd.Configure()
	if err != nil {
		return &TestError{ error: fmt.Errorf("sdcard configure error: %s", err.Error()), Code: 2 }
	}

	filesystem := fatfs.New(&sd)

	// Configure FATFS with sector size (must match value in ff.h - use 512)
	filesystem.Configure(&fatfs.Config{
		SectorSize: 512,
	})

	err = filesystem.Mount()
	if err != nil {
		return &TestError{ error: fmt.Errorf("mount error: %s", err.Error()), Code: 3 }
	}
	fmt.Printf("card mount ok\r\n")

	var volumeAtt uint8 = 60
	musicPlayer := vs1053.NewPlayer(&codec)
	musicPlayer.SetVolume(volumeAtt, volumeAtt)

	var f tinyfs.File
	var ff *fatfs.File

	fmt.Printf("Playing track 001 (by Blocking)\r\n");
	f, _ = filesystem.OpenFile("/track001.mp3", os.O_RDONLY)
	ff, _ = f.(*fatfs.File)
	// Play one file, don't return until complete
	musicPlayer.PlayFullFile(ff);
	ff.Close()

	fmt.Printf("Playing track 002 (by Non-Blocking)\r\n");
	f, _ = filesystem.OpenFile("/track002.mp3", os.O_RDONLY)
	ff, _ = f.(*fatfs.File)
	// Play another file in the background
	musicPlayer.StartPlayingFile(ff);

	// file is playing in the background
	for loop := 0; ; loop++ {
		if musicPlayer.Stopped() {
			fmt.Printf("Done playing music\r\n")
			return nil
		}
		if serial.Buffered() > 0 {
			data, _ := serial.ReadByte()
			switch data {
			case 's':
				musicPlayer.StopPlaying()
			case 'p':
				if !musicPlayer.Paused() {
					fmt.Printf("Paused\r\n")
					musicPlayer.PausePlaying(true)
				} else {
					fmt.Printf("Resumed\r\n")
					musicPlayer.PausePlaying(false)
				}
			case '=':
				fallthrough
			case '+':
				if volumeAtt > 0 {
					volumeAtt--
					musicPlayer.SetVolume(volumeAtt, volumeAtt)
				}
			case '-':
				if volumeAtt < 254 {
					volumeAtt++
					musicPlayer.SetVolume(volumeAtt, volumeAtt)
				}
			default:
			}
		}
		if loop % 10 == 0 {
			led.Toggle()
		}
		time.Sleep(10 * time.Millisecond)
	}

	ff.Close()

	return nil
}