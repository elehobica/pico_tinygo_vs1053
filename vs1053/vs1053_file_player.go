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
    "fmt"
    "io"
    "time"
)

type File interface {
    Tell() (ret int64, err error)
    Seek(offset int64) error
    Read(buf []byte) (n int, err error)
}

type Player struct {
    codec        *Device
    isPlaying    bool
    isPaused     bool
    currentTrack File
    mp3Buf       []byte
    mp3BufReq    chan struct{}
}

const (
    DATA_BUF_LEN uint32 = 32 //!< Length of the data buffer
    REQ_CH_SZ    uint32 = 1  //!< Size of the request channel
)

func NewPlayer(codec *Device) Player {
    buff := make([]byte, DATA_BUF_LEN)
    return Player{
        codec:        codec,
        isPlaying:    false,
        isPaused:     false,
        currentTrack: nil,
        mp3Buf:       buff,
        mp3BufReq:    nil,
    }
}

func (p *Player) PlayFullFile(file File) error {
    err := p.StartPlayingFile(file)
    if err != nil {
        return fmt.Errorf("StartPlayingFile failed")
    }

    for p.isPlaying && !p.isPaused {
        time.Sleep(10 * time.Millisecond) // give goroutine a chance to run
    }

    // music file finished!
    return nil
}

func (p *Player) StopPlaying() error {
    // stop DreqInterrupt and cancel all playback
    p.codec.setDreqInterrupt(false, nil)
    p.codec.sciWrite(REG_MODE, MODE_SM_LINE1 | MODE_SM_SDINEW | MODE_SM_CANCEL)

    // wrap it up!
    close(p.mp3BufReq)
    return nil
}

func (p *Player) PausePlaying(pause bool) error {
    p.isPaused = pause
    if p.isPlaying && !p.isPaused {
        p.feedBuffer()
    }
    return nil
}

func (p *Player) Paused() bool {
    return p.isPlaying && p.isPaused
}

func (p *Player) Stopped() bool {
    return !p.isPlaying
}

func (p *Player) SetVolume(left, right uint8) {
    p.codec.SetVolume(left, right)
}

func (p *Player) StartPlayingFile(file File) error {
    // reset playback
    p.codec.sciWrite(REG_MODE, MODE_SM_LINE1 | MODE_SM_SDINEW | MODE_SM_LAYER12)

    // resync
    p.codec.sciWrite(REG_WRAMADDR, 0x1e29)
    p.codec.sciWrite(REG_WRAM, 0)

    p.currentTrack = file

    // We know we have a valid file. Check if .mp3
    // If so, check for ID3 tag and jump it if present.
    pos, err := p.mp3_ID3Jumper(p.currentTrack)
    if err != nil {
        return fmt.Errorf("mp3_ID3Jumper failed: %s", err.Error())
    }
    p.currentTrack.Seek(pos)

    // As explained in datasheet, set twice 0 in REG_DECODETIME to set time back to 0
    p.codec.sciWrite(REG_DECODETIME, 0x00)
    p.codec.sciWrite(REG_DECODETIME, 0x00)

    p.isPlaying = true
    p.isPaused = false

    // wait till its ready for data
    for !p.codec.readyForData() {}

    // fill it up!
    for p.isPlaying && !p.isPaused && p.codec.readyForData() {
        p.feedBuffer()
    }

    // open channel & set interrupt
    p.mp3BufReq = make(chan struct{}, REQ_CH_SZ)
    p.codec.setDreqInterrupt(true, func() {
        p.mp3BufReq <- struct{}{} // send event (no type)
    })

    // ok going forward, we can use goroutine
    go func(req <-chan struct{}) {
        for {
            _, more := <-req
            if !more {
                p.isPlaying = false
                p.isPaused = false
                p.codec.setDreqInterrupt(false, nil)
                return
            }
            p.feedBuffer()
        }
    } (p.mp3BufReq)

    return nil
}

func (p *Player) mp3_ID3Jumper(mp3 File) (start int64, err error) {
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

func (p *Player) feedBuffer() {
    if !p.isPlaying || p.isPaused || !p.codec.readyForData() {
       return // paused or stopped
    }

    // Feed the hungry buffer! :)
    for p.codec.readyForData() {
        // Read some audio data from the SD card file
        br, err := p.currentTrack.Read(p.mp3Buf)

        if err == io.EOF {
            // must be at the end of the file, wrap it up!
            close(p.mp3BufReq)
            break
        }

        p.codec.playData(p.mp3Buf[:br])
    }
}
