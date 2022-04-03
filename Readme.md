# Raspberry Pi Pico TinyGo tinyfs Test
## Overview
This project is an example of tinyfs on Raspberry Pi Pico by TinyGo.
This project supports:
* TinyGo 0.22.0
* SD card access by SPI interface
* FAT16, FAT32 formats
* write/read speed benchmark

## Supported Board
* Raspberry Pi Pico

## Ciruit Diagram
![Circuit Diagram](doc/Pico_FatFs_Test_Schematic.png)

## Pin Assignment
### microSD card

| Pico Pin # | Pin Name | Function | microSD connector | microSD SPI board |
----|----|----|----|----
|  4 | GP2 | SPI0_SCK | CLK (5) | CLK |
|  5 | GP3 | SPI0_TX | CMD (3) | MOSI |
|  6 | GP4 | SPI0_RX | DAT0 (7) | MISO |
|  7 | GP5 | SPI0_CSn | CD/DAT3 (2) | CS |
|  8 | GND | GND | VSS (6) | GND |
| 36 | 3V3(OUT) | 3.3V | VDD (4) | 3V3 |

#### Caution
* SPI0_TX and SPI0_RX needs to be pull-ed up with 10Kohm.
* Wire length between Pico and SD card is very sensitive. Short wiring as possible is desired, otherwise errors such as Mount error, Preallocation error and Write fail will occur.

### Serial (CP2102 module)
| Pico Pin # | Pin Name | Function | CP2102 module |
----|----|----|----
|  1 | GP0 | UART0_TX | RXD |
|  2 | GP1 | UART0_RX | TXD |
|  3 | GND | GND | GND |

## How to build
* Build is confirmed only in TinyGo Docker environment with Windows WSL2 integration
* Before starting docker, clone repository to your local enviroment (by GitBash etc.)
```
> cd /d/somewhere/share
> git clone -b main https://github.com/elehobica/pico_tinyfs_test.git
```

* Docker
```
> wsl
(in WSL2 shell)
$ docker pull docker pull tinygo/tinygo
$ docker images
$ docker run -it -v /mnt/d/somewhere/share:/share tinygo/tinygo:latest /bin/bash
(in docker container)
# cd /share

(copy repository for docker native directory for best performance of WSL2, otherwise stay /share)
(# cp -r /share/pico_tinyfs_test ~/ && cd ~ )

# cd pico_tinyfs_test
```

* Go Module Configuration
```
# go mod init pico_tinyfs_test
# cd console
# go mod init mylocal.com/console
# cd ..
# go mod edit -replace mylocal.com/console=./console
# go mod tidy
```

* TinyGo Build
```
# tinygo build -target=pico -o pico_tinyfs_test.uf2

(copy UF2 back to Windows local if working on docker native directory)
(# cp pico_tinyfs_test.uf2 /share/pico_tinyfs_test/ )
```

* Put UF2 

Then, go back to Windows environment and put "pico_tinyfs_test.uf2" on RPI-RP2 drive