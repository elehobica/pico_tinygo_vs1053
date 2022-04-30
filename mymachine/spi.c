#include "spi.h"

// ----------------------------------------------------------------------------
// Generic input/output

/*! \brief Check whether a write can be done on SPI device
 *  \ingroup hardware_spi
 *
 * \param spi SPI instance specifier, either \ref spi0 or \ref spi1
 * \return false if no space is available to write. True if a write is possible
 */
static inline bool spi_is_writable(const spi_inst_t *spi) {
    return (spi_get_const_hw(spi)->sr & SPI_SSPSR_TNF_BITS);
}

/*! \brief Check whether a read can be done on SPI device
 *  \ingroup hardware_spi
 *
 * \param spi SPI instance specifier, either \ref spi0 or \ref spi1
 * \return true if a read is possible i.e. data is present
 */
static inline bool spi_is_readable(const spi_inst_t *spi) {
    return (spi_get_const_hw(spi)->sr & SPI_SSPSR_RNE_BITS);
}

// Write len bytes from src to SPI. Simultaneously read len bytes from SPI to dst.
// Note this function is guaranteed to exit in a known amount of time (bits sent * time per bit)
int spi_write_read_blocking(spi_inst_t *spi, const uint8_t *src, uint8_t *dst, size_t len) {
    invalid_params_if(SPI, 0 > (int)len);

    // Never have more transfers in flight than will fit into the RX FIFO,
    // else FIFO will overflow if this code is heavily interrupted.
    const size_t fifo_depth = 8;
    size_t rx_remaining = len, tx_remaining = len;

    while (rx_remaining || tx_remaining) {
        if (tx_remaining && spi_is_writable(spi) && rx_remaining < tx_remaining + fifo_depth) {
            spi_get_hw(spi)->dr = (uint32_t) *src++;
            --tx_remaining;
        }
        if (rx_remaining && spi_is_readable(spi)) {
            *dst++ = (uint8_t) spi_get_hw(spi)->dr;
            --rx_remaining;
        }
    }

    return (int)len;
}

// Write len bytes directly from src to the SPI, and discard any data received back
int spi_write_blocking(spi_inst_t *spi, const uint8_t *src, size_t len) {
    invalid_params_if(SPI, 0 > (int)len);
    // Write to TX FIFO whilst ignoring RX, then clean up afterward. When RX
    // is full, PL022 inhibits RX pushes, and sets a sticky flag on
    // push-on-full, but continues shifting. Safe if SSPIMSC_RORIM is not set.
    for (size_t i = 0; i < len; ++i) {
        while (!spi_is_writable(spi))
            tight_loop_contents();
        spi_get_hw(spi)->dr = (uint32_t)src[i];
    }
    // Drain RX FIFO, then wait for shifting to finish (which may be *after*
    // TX FIFO drains), then drain RX FIFO again
    while (spi_is_readable(spi))
        (void)spi_get_hw(spi)->dr;
    while (spi_get_hw(spi)->sr & SPI_SSPSR_BSY_BITS)
        tight_loop_contents();
    while (spi_is_readable(spi))
        (void)spi_get_hw(spi)->dr;

    // Don't leave overrun flag set
    spi_get_hw(spi)->icr = SPI_SSPICR_RORIC_BITS;

    return (int)len;
}

// Read len bytes directly from the SPI to dst.
// repeated_tx_data is output repeatedly on SO as data is read in from SI.
// Generally this can be 0, but some devices require a specific value here,
// e.g. SD cards expect 0xff
int spi_read_blocking(spi_inst_t *spi, uint8_t repeated_tx_data, uint8_t *dst, size_t len) {
    invalid_params_if(SPI, 0 > (int)len);
    const size_t fifo_depth = 8;
    size_t rx_remaining = len, tx_remaining = len;

    while (rx_remaining || tx_remaining) {
        if (tx_remaining && spi_is_writable(spi) && rx_remaining < tx_remaining + fifo_depth) {
            spi_get_hw(spi)->dr = (uint32_t) repeated_tx_data;
            --tx_remaining;
        }
        if (rx_remaining && spi_is_readable(spi)) {
            *dst++ = (uint8_t) spi_get_hw(spi)->dr;
            --rx_remaining;
        }
    }

    return (int)len;
}