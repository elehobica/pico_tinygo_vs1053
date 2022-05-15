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
    "time"
    "sync"
    "fmt"
	"github.com/elehobica/pico_tinygo_vs1053/mymachine"
)

type Device struct {
	bus        mymachine.SPI
    spiMutex   sync.Mutex
    csPin      machine.Pin
    sckPin     machine.Pin
    mosiPin    machine.Pin
    misoPin    machine.Pin
    rstPin     machine.Pin
    dcsPin     machine.Pin
    dreqPin    machine.Pin
    slowBaud   *mymachine.SPIBaudRateReg
    fastBaud   *mymachine.SPIBaudRateReg
}

const (
    SlowFreq = 1000000 // below 12.288 MHz * 3.0 / 7 (for SCI Read)
    FastFreq = 8000000 // below 12.288 MHz * 3.0 / 4 (for SCI Write and SDI Write)
)

func New(bus mymachine.SPI, sckPin, mosiPin, misoPin, csPin, rstPin, dcsPin, dreqPin machine.Pin) Device {
    return Device{
        bus:        bus,
        spiMutex:   sync.Mutex{},
        sckPin:     sckPin,
        mosiPin:    mosiPin,
        misoPin:    misoPin,
        csPin:      csPin,
        rstPin:     rstPin,
        dcsPin:     dcsPin,
        dreqPin:    dreqPin,
        slowBaud:   nil,
        fastBaud:   nil,
    }
}

func (d *Device) Configure() error {
    version := d.begin()
    if version != VER_VS1053 {
        return fmt.Errorf("vs1053 version: %d is not %d", version, VER_VS1053)
    }
    return nil
}

func (d *Device) softReset() {
    d.sciWrite(REG_MODE, MODE_SM_SDINEW | MODE_SM_RESET)
    time.Sleep(100 * time.Millisecond)
}

func (d *Device) reset() {
    // TODO:
    // http://www.vlsi.fi/player_vs1011_1002_1003/modularplayer/vs10xx_8c.html#a3
    // hardware reset
    if d.rstPin != machine.NoPin {
        d.rstPin.Low()
        time.Sleep(100 * time.Millisecond)
        d.rstPin.High()
    }
    d.csPin.High()
    d.dcsPin.High()
    time.Sleep(100 * time.Millisecond)
    d.softReset()
    time.Sleep(100 * time.Millisecond)

    // CLOCKF
    //  b15-13: SC_MULT (multiply XTALI): 0: x1.0, 1: x2.0, 2: x2.5, 3: x3.0, 4: x3.5, 5: x4.0, 6: x4.5, 7: x5.0
    //  b12-11: SC_ADD  (f/w multiplier): 0: no modification, 1: x1.0, 2: x1.5, 3: x2.0
    //  b10: 0: SC_FREQ: 0 when 12.288 MHz operation
    d.sciWrite(REG_CLOCKF, 0x6000) // x3.0

    d.SetVolume(40, 40)
}

func (d *Device) begin() (version uint8) {
    if d.rstPin != machine.NoPin {
        d.rstPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
        d.rstPin.Low()
    }

    d.csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
    d.csPin.High()
    d.dcsPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
    d.dcsPin.High()
    d.dreqPin.Configure(machine.PinConfig{Mode: machine.PinInput})

    d.bus.Configure(machine.SPIConfig{
        SCK:       d.sckPin,
        SDO:       d.mosiPin,
        SDI:       d.misoPin,
        Frequency: SlowFreq,
        LSBFirst:  false,
        Mode:      0, // phase=0, polarity=0
    })

    // save SPI BaudRate for SCI and SDI
    d.slowBaud, _ = d.bus.SaveBaudRate(SlowFreq)
    d.fastBaud, _ = d.bus.SaveBaudRate(FastFreq)

    d.reset()

    status := uint8(d.sciRead(REG_STATUS))
    version = (status >> 4) & 0x0F
    return version
}

func (d *Device) SwitchToMp3Mode() {
    d.sciWrite(REG_WRAMADDR, 0xc017)
    d.sciWrite(REG_WRAM, 3)
    d.sciWrite(REG_WRAMADDR, 0xc019)
    d.sciWrite(REG_WRAM, 3)
    time.Sleep(100 * time.Millisecond)
    d.softReset()
}

func (d *Device) SetVolume(left, right uint8) {
    // accepts values between 0 and 255 for left and right.
    // maximum volume is 0x0000 and total silence is 0xFEFE.
    // Setting SCI_VOL to 0xFFFF will activate analog powerdown mode.
    var v uint16 = (uint16(left) << 8) | uint16(right)

    d.sciWrite(REG_VOLUME, v)
}

func (d *Device) sciRead(addr uint8) (data uint16) {
    d.spiMutex.Lock()
    defer d.spiMutex.Unlock()
	d.csPin.Low()
	defer d.csPin.High()
    d.bus.RestoreBaudRate(d.slowBaud)
    d.bus.Transfer(SCI_READ)
    d.bus.Transfer(addr)
    time.Sleep(10 * time.Microsecond)
    data0, _ := d.bus.Transfer(0x00)
    data1, _ := d.bus.Transfer(0x00)
    data =(uint16(data0) << 8) | uint16(data1)
    return data
}

func (d *Device) sciWrite(addr uint8, data uint16) {
    d.spiMutex.Lock()
    defer d.spiMutex.Unlock()
	d.csPin.Low()
    defer d.csPin.High()
    d.bus.RestoreBaudRate(d.fastBaud)
    d.bus.Transfer(SCI_WRITE)
    d.bus.Transfer(addr)
    d.bus.Transfer(uint8(data >> 8))
    d.bus.Transfer(uint8(data & 0xff))
}

func (d *Device) readyForData() bool {
    return d.dreqPin.Get()
}

func (d *Device) playData(buf []byte) {
    d.spiMutex.Lock()
    defer d.spiMutex.Unlock()
	d.dcsPin.Low()
	defer d.dcsPin.High()
    d.bus.RestoreBaudRate(d.fastBaud)
    d.bus.Tx(buf, nil)
}

func (d *Device) setDreqInterrupt(flg bool, f func()) error {
    if flg {
        if d.dreqPin == machine.NoPin {
            return fmt.Errorf("vs1053 failed to set interrupt")
        }
        d.dreqPin.SetInterrupt(machine.PinRising, func (pin machine.Pin) { f() })
    } else {
        d.dreqPin.SetInterrupt(machine.PinRising, nil)
    }
    return nil
}