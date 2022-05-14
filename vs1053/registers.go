package vs1053

// Commands
const (
    SCI_READ  = 0x03 //!< Serial read address
    SCI_WRITE = 0x02 //!< Serial write address
)

// Registers
const (
    REG_MODE       = 0x00 //!< Mode control
    REG_STATUS     = 0x01 //!< Status of VS1053b
    REG_BASS       = 0x02 //!< Built-in bass/treble control
    REG_CLOCKF     = 0x03 //!< Clock frequency + multiplier
    REG_DECODETIME = 0x04 //!< Decode time in seconds
    REG_AUDATA     = 0x05 //!< Misc. audio data
    REG_WRAM       = 0x06 //!< RAM write/read
    REG_WRAMADDR   = 0x07 //!< Base address for RAM write/read
    REG_HDAT0      = 0x08 //!< Stream header data 0
    REG_HDAT1      = 0x09 //!< Stream header data 1
    REG_VOLUME     = 0x0B //!< Volume control
)

// Register Values
const (
    INT_ENABLE       = 0xC01A //!< Interrupt enable
    MODE_SM_DIFF     = 0x0001 //!< Differential, 0: normal in-phase audio, 1: left channel inverted
    MODE_SM_LAYER12  = 0x0002 //!< Allow MPEG layers I & II
    MODE_SM_RESET    = 0x0004 //!< Soft reset
    MODE_SM_CANCEL   = 0x0008 //!< Cancel decoding current file
    MODE_SM_EARSPKLO = 0x0010 //!< EarSpeaker low setting
    MODE_SM_TESTS    = 0x0020 //!< Allow SDI tests
    MODE_SM_STREAM   = 0x0040 //!< Stream mode
    MODE_SM_SDINEW   = 0x0800 //!< VS1002 native SPI modes
    MODE_SM_ADPCM    = 0x1000 //!< PCM/ADPCM recording active
    MODE_SM_LINE1    = 0x4000 //!< MIC/LINE1 selector, 0: MICP, 1: LINE1
    MODE_SM_CLKRANGE = 0x8000 //!< Input clock range, 0: 12..13 MHz, 1: 24..26 MHz
    SCI_AIADDR       = 0x0A //!< Indicates the start address of the application code written earlier
                            //!< with SCI_WRAMADDR and SCI_WRAM registers.
    SCI_AICTRL0      = 0x0C //!< SCI_AICTRL register 0. Used to access the user's application program
    SCI_AICTRL1      = 0x0D //!< SCI_AICTRL register 1. Used to access the user's application program
    SCI_AICTRL2      = 0x0E //!< SCI_AICTRL register 2. Used to access the user's application program
    SCI_AICTRL3      = 0x0F //!< SCI_AICTRL register 3. Used to access the user's application program
    // SS_VER of SCI_STATUS[7:4]
    VER_VS1001 = 0x00
    VER_VS1011 = 0x01
    VER_VS1002 = 0x02
    VER_VS1003 = 0x03
    VER_VS1053 = 0x04
    VER_VS8053 = 0x04
    VER_VS1033 = 0x05
    VER_VS1063 = 0x06
    VER_VS1103 = 0x07
)