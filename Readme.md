# Raspberry Pi Pico TinyGo FatFs Test
## Overview
This project is an example of FatFs on Raspberry Pi Pico by TinyGo.
* ported from [elehobica/pico_fatfs_test](https://github.com/elehobica/pico_fatfs_test)
* confirmed with TinyGo 0.22.0

This project supports:
* FatFs R0.13c ([http://elm-chan.org/fsw/ff/00index_e.html](http://elm-chan.org/fsw/ff/00index_e.html))
* SD card access by SPI interface
* SD, SDHC, SDXC cards
* FAT16, FAT32, exFAT formats
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
> git clone -b main https://github.com/elehobica/pico_tinygo_fatfs_test.git
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
(# cp -r /share/pico_tinygo_fatfs_test ~/ && cd ~ )

# cd pico_tinygo_fatfs_test
```

* Go Module Configuration
```
# go mod init pico_tinygo_fatfs_test
# go mod tidy
```

* TinyGo Build
```
# tinygo build -target=pico -o pico_tinygo_fatfs_test.uf2

(copy UF2 back to Windows local if working on docker native directory)
(# cp pico_tinygo_fatfs_test.uf2 /share/pico_tinygo_fatfs_test/ )
```

* Put UF2 

Then, go back to Windows environment and put "pico_tinygo_fatfs_test.uf2" on RPI-RP2 drive

## Benchmark Comparison with C++
### Sansung microSDHC EVO Plus 32GB (UHS-I U1)
* Reference [pico_fatfs_test](https://github.com/elehobica/pico_fatfs_test)
```
=====================
== pico_fatfs_test ==
=====================
mount ok
Type is FAT32
Card size:   32.00 GB (GB = 1E9 bytes)

FILE_SIZE_MB = 5
BUF_SIZE = 512 bytes
Starting write test, please wait.

write speed and latency
speed,max,min,avg
KB/Sec,usec,usec,usec
447.7192, 6896, 1007, 1142
446.4797, 7589, 1024, 1145

Starting read test, please wait.

read speed and latency
speed,max,min,avg
KB/Sec,usec,usec,usec
974.9766, 1050, 403, 524
974.4066, 1049, 402, 524
```

* TinyGo (this repository)
```
============================
== pico_tinygo_fatfs_test ==
============================
mount ok
Type is FAT32
Card size:   32.00 GB (GB = 1E9 bytes)

FILE_SIZE_MB = 5
BUF_SIZE = 512 bytes
Starting write test, please wait.

write speed and latency
speed,max,min,avg
KB/Sec,usec,usec,usec
349.1884, 17431, 1042, 1442
361.0920, 22377, 1062, 1391

Starting read test, please wait.

read speed and latency
speed,max,min,avg
KB/Sec,usec,usec,usec
354.7382, 22428, 1284, 1417
353.7843, 22470, 1293, 1421
```