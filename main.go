package main

import (
	"fmt"
	"machine"
	"time"

	//"tinygo.org/x/drivers/sdcard"
	"github.com/elehobica/pico_tinygo_vs1053/sdcard"
	//"tinygo.org/x/tinyfs/fatfs"
	"github.com/elehobica/pico_tinygo_vs1053/fatfs"
	"github.com/elehobica/pico_tinygo_vs1053/mymachine"
	"github.com/elehobica/pico_tinygo_vs1053/vs1053"
)

var (
	spi0     mymachine.SPI
	sckPin   machine.Pin
	sdoPin   machine.Pin
	sdiPin   machine.Pin
	csPin    machine.Pin

	spi1     mymachine.SPI
	sck1Pin  machine.Pin
	sdo1Pin  machine.Pin
	sdi1Pin  machine.Pin
	cs1Pin   machine.Pin
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

	err := vs1053_test(led)
	if err != nil {
		fmt.Printf("ERROR[%d]: %s\r\n", err.Code, err.Error())
		led.ErrorBlinkFor(err.Code)
	}

	led.OkBlinkFor()
}

func vs1053_test(led *Pin) (testError *TestError) {
	println(); println()
	println("========================")
	println("== pico_tinygo_vs1053 ==")
	println("========================")

	codec := vs1053.New(spi1, sck1Pin, sdo1Pin, sdi1Pin,  cs1Pin, xrstPin, xdcsPin, xdreqPin)
	err := codec.Configure()
	if err != nil {
		return &TestError{ error: fmt.Errorf("codec configure error: %s", err.Error()), Code: 1 }
	}

	sd := sdcard.New(spi0, sckPin, sdoPin, sdiPin, csPin)
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

	codec.SwitchToMp3Mode()

	var volumeAtt uint8 = 60
	musicPlayer := vs1053.NewPlayer(&codec, filesystem)
	musicPlayer.SetVolume(volumeAtt, volumeAtt)

	// Play one file, don't return until complete
	fmt.Printf("Playing track 001 (by Blocking)\r\n");
	musicPlayer.PlayFullFile("/track001.mp3");

	// If DREQ is on an interrupt pin, we can do background audio playing
	musicPlayer.UseInterrupt()

	// Play another file in the background, REQUIRES interrupts!
	fmt.Printf("Playing track 002 (by Interrupt)\r\n");
	musicPlayer.StartPlayingFile("/track002.mp3");

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

	return nil
}