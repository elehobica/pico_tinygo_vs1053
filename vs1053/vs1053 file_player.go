/*!
 * @file Adafruit_VS1053.cpp
 *
 * @mainpage Adafruit VS1053 Library
 *
 * @section intro_sec Introduction
 *
 * This is a library for the Adafruit VS1053 Codec Breakout
 *
 * Designed specifically to work with the Adafruit VS1053 Codec Breakout
 * ----> https://www.adafruit.com/products/1381
 *
 * Adafruit invests time and resources providing this open source code,
 * please support Adafruit and open-source hardware by purchasing
 * products from Adafruit!
 *
 * @section author Author
 *
 * Written by Limor Fried/Ladyada for Adafruit Industries.
 *
 * @section license License
 *
 * BSD license, all text above must be included in any redistribution
 */

/*
 * ported to TinyGo by Elehobica, 2022
 */

package vs1053

import (
    "machine"
    "fmt"
	"io"
	"os"
    "time"
    "sync"
    "strings"
	"github.com/elehobica/pico_tinygo_vs1053/fatfs"
)

type Player struct {
    codec        *Device
    fs           *fatfs.FATFS
    mu           sync.Mutex
    playingMusic bool
	currentTrack *fatfs.File
    mp3buffer    []byte
}

var (
    DATABUFFERLEN uint32 = 32 //!< Length of the data buffer
)

func NewPlayer(codec *Device, fs *fatfs.FATFS) Player {
	buff := make([]byte, DATABUFFERLEN)
    return Player{
        codec:        codec,
        fs:           fs,
        mu:           sync.Mutex{},
        playingMusic: false,
        currentTrack: nil,
        mp3buffer:    buff,
    }
}

func (p *Player) PlayFullFile(trackname string) error {
    err := p.StartPlayingFile(trackname)
    if err != nil {
        return fmt.Errorf("StartPlayingFile failed")
    }
    for p.playingMusic {
        p.feedBuffer()
        time.Sleep(5 * time.Millisecond) // give IRQs a chance
    }
    // music file finished!
    return nil
}

func (p *Player) StopPlaying() error {
    // cancel all playback
    p.codec.sciWrite(REG_MODE, MODE_SM_LINE1 | MODE_SM_SDINEW | MODE_SM_CANCEL)

    // wrap it up!
    p.playingMusic = false
	p.currentTrack.Close()
	p.currentTrack = nil
    return nil
}

func (p *Player) PausePlaying(pause bool) error {
    p.playingMusic = !pause && p.currentTrack != nil
    if p.playingMusic {
        p.feedBuffer()
    }
    return nil
}

func (p *Player) Paused() bool {
    return !p.playingMusic && p.currentTrack != nil
}

func (p *Player) Stopped() bool {
    return !p.playingMusic && p.currentTrack == nil
}

func (p *Player) SetVolume(left, right uint8) {
    p.codec.SetVolume(left, right)
}

func (p *Player) UseInterrupt() error {
    if p.codec.dreqPin == machine.NoPin {
        return fmt.Errorf("vs1053_file_player failed to use interrupt")
    }
    p.codec.dreqPin.SetInterrupt(machine.PinRising, p.feedBufferIRQ)
    return nil
}

func (p *Player) StartPlayingFile(file string) error {
    // reset playback
    p.codec.sciWrite(REG_MODE, MODE_SM_LINE1 | MODE_SM_SDINEW | MODE_SM_LAYER12)

    // resync
    p.codec.sciWrite(REG_WRAMADDR, 0x1e29)
    p.codec.sciWrite(REG_WRAM, 0)

	f, err := p.fs.OpenFile(file, os.O_RDONLY)
	if err != nil {
		return fmt.Errorf("open error: %s", err.Error())
	}

	ff, ok := f.(*fatfs.File)
	if ok != true {
        return fmt.Errorf("conversion to *fatfs.File failed")
    }
	p.currentTrack = ff

    // We know we have a valid file. Check if .mp3
    // If so, check for ID3 tag and jump it if present.
    if p.isMp3File(file) {
        pos, err := p.mp3_ID3Jumper(p.currentTrack)
        if err != nil {
            return fmt.Errorf("mp3_ID3Jumper failed: %s", err.Error())
        }
        p.currentTrack.Seek(pos)
    }

    // don't let the IRQ get triggered by accident here
    p.codec.noInterrupts()
    defer p.codec.interrupts()

    // As explained in datasheet, set twice 0 in REG_DECODETIME to set time back to 0
    p.codec.sciWrite(REG_DECODETIME, 0x00)
    p.codec.sciWrite(REG_DECODETIME, 0x00)

    p.playingMusic = true

    // wait till its ready for data
    for !p.readyForData() {}

    // fill it up!
    for p.playingMusic && p.readyForData() {
        p.feedBuffer()
    }

    // ok going forward, we can use the IRQ

    return nil
}

// Just checks to see if the name ends in ".mp3"
func (p *Player) isMp3File(file string) bool {
    return strings.HasSuffix(file, ".mp3")
}

func (p *Player) mp3_ID3Jumper(mp3 *fatfs.File) (start int64, err error) {
    start = 0
    if mp3 == nil {
        return 0, fmt.Errorf("nil file")
    }
    current, _ := mp3.Tell()
    err = mp3.Seek(0)
    if err != nil {
        return 0, fmt.Errorf("Seek failed")
    }
    defer mp3.Seek(current)
    buf := make([]byte, 3)
    br, err := mp3.Read(buf)
    if err != nil || br != 3 {
        return 0, fmt.Errorf("Read failed")
    }
    if string(buf) == "ID3" {
        mp3.Seek(6)
        for i := 0; i < 4; i++ {
            start <<= 7
            mp3.Read(buf[:1])
            start |= 0x7F & int64(buf[0])
        }
    } else {
        return 0, fmt.Errorf("It wasn't the damn TAG.")
    }
    return start, nil
}

func (p *Player) feedBufferIRQ(pin machine.Pin) {
    // Here pin == machine.NoPin
    p.feedBuffer()
}

func (p *Player) feedBuffer() {
    p.codec.noInterrupts()
    defer p.codec.interrupts()
    p.mu.Lock()
    defer p.mu.Unlock()

    if !p.playingMusic || p.currentTrack == nil || !p.readyForData() {
       return // paused or stopped
    }

    // Feed the hungry buffer! :)
    for p.readyForData() {
        // Read some audio data from the SD card file
        bytesread, err := p.currentTrack.Read(p.mp3buffer)

        if err == io.EOF {
            // must be at the end of the file, wrap it up!
            p.playingMusic = false
            p.currentTrack.Close()
            p.currentTrack = nil
            break
        }

        p.playData(bytesread)
    }
}

func (p *Player) readyForData() bool {
    return p.codec.dreqPin.Get()
}

func (p *Player) playData(n int) {
	p.codec.dcsPin.Low()
	defer p.codec.dcsPin.High()
    p.codec.bus.Tx(p.mp3buffer[:n], nil)
}