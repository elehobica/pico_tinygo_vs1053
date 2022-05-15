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
	spi0     *machine.SPI
	sckPin   machine.Pin
	sdoPin   machine.Pin
	sdiPin   machine.Pin
	csPin    machine.Pin

	spi1     *machine.SPI
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

	mySpi0 := mymachine.NewSPI(spi0)
	mySpi0.Configure(machine.SPIConfig{
		SCK:       sckPin,
		SDO:       sdoPin,
		SDI:       sdiPin,
		LSBFirst:  false,
		Mode:      0, // phase=0, polarity=0
	})

	mySpi1 := mymachine.NewSPI(spi1)
	mySpi1.Configure(machine.SPIConfig{
		SCK:       sck1Pin,
		SDO:       sdo1Pin,
		SDI:       sdi1Pin,
		LSBFirst:  false,
		Mode:      0, // phase=0, polarity=0
	})

	err := vs1053_test(led, mySpi0, mySpi1)
	if err != nil {
		fmt.Printf("ERROR[%d]: %s\r\n", err.Code, err.Error())
		led.ErrorBlinkFor(err.Code)
	}

	led.OkBlinkFor()
}

func vs1053_test(led *Pin, spi0 mymachine.SPI, spi1 mymachine.SPI) (testError *TestError) {
	println(); println()
	println("========================")
	println("== pico_tinygo_vs1053 ==")
	println("========================")

	codec := vs1053.New(spi1, cs1Pin, xrstPin, xdcsPin, xdreqPin)
	err := codec.Configure()
	if err != nil {
		return &TestError{ error: fmt.Errorf("codec configure error: %s", err.Error()), Code: 1 }
	}
	codec.SwitchToMp3Mode()

	sd := sdcard.New(spi0, csPin)
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
	musicPlayer := vs1053.NewPlayer(&codec, filesystem)
	musicPlayer.SetVolume(volumeAtt, volumeAtt)

	// Play one file, don't return until complete
	fmt.Printf("Playing track 001 (by Blocking)\r\n");
	musicPlayer.PlayFullFile("/track001.mp3");

	// Play another file in the background
	fmt.Printf("Playing track 002 (by Non-Blocking)\r\n");
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