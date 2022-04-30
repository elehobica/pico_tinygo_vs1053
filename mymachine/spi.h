/*
 * Copyright (c) 2020 Raspberry Pi (Trading) Ltd.
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

#ifndef _HARDWARE_SPI_H
#define _HARDWARE_SPI_H

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <stddef.h>

typedef unsigned int uint;

#include "pico_assert.h"
#include "hardware/structs/spi.h"

// PICO_CONFIG: PARAM_ASSERTIONS_ENABLED_SPI, Enable/disable assertions in the SPI module, type=bool, default=0, group=hardware_spi
#ifndef PARAM_ASSERTIONS_ENABLED_SPI
#define PARAM_ASSERTIONS_ENABLED_SPI 0
#endif

#ifdef __cplusplus
extern "C" {
#endif

/*! \brief No-op function for the body of tight loops
 *  \ingroup pico_platform
 *
 * Np-op function intended to be called by any tight hardware polling loop. Using this ubiquitously
 * makes it much easier to find tight loops, but also in the future \#ifdef-ed support for lockup
 * debugging might be added
 */
static inline void tight_loop_contents(void) {}

/** \file hardware/spi.h
 *  \defgroup hardware_spi hardware_spi
 *
 * Hardware SPI API
 *
 * RP2040 has 2 identical instances of the Serial Peripheral Interface (SPI) controller.
 *
 * The PrimeCell SSP is a master or slave interface for synchronous serial communication with peripheral devices that have
 * Motorola SPI, National Semiconductor Microwire, or Texas Instruments synchronous serial interfaces.
 *
 * Controller can be defined as master or slave using the \ref spi_set_slave function.
 *
 * Each controller can be connected to a number of GPIO pins, see the datasheet GPIO function selection table for more information.
 */

// PICO_CONFIG: PICO_DEFAULT_SPI, Define the default SPI for a board, min=0, max=1, group=hardware_spi
// PICO_CONFIG: PICO_DEFAULT_SPI_SCK_PIN, Define the default SPI SCK pin, min=0, max=29, group=hardware_spi
// PICO_CONFIG: PICO_DEFAULT_SPI_TX_PIN, Define the default SPI TX pin, min=0, max=29, group=hardware_spi
// PICO_CONFIG: PICO_DEFAULT_SPI_RX_PIN, Define the default SPI RX pin, min=0, max=29, group=hardware_spi
// PICO_CONFIG: PICO_DEFAULT_SPI_CSN_PIN, Define the default SPI CSN pin, min=0, max=29, group=hardware_spi

/**
 * Opaque type representing an SPI instance.
 */
typedef struct spi_inst spi_inst_t;

/** Identifier for the first (SPI 0) hardware SPI instance (for use in SPI functions).
 *
 * e.g. spi_init(spi0, 48000)
 *
 *  \ingroup hardware_spi
 */
#define spi0 ((spi_inst_t * const)spi0_hw)

/** Identifier for the second (SPI 1) hardware SPI instance (for use in SPI functions).
 *
 * e.g. spi_init(spi1, 48000)
 *
 *  \ingroup hardware_spi
 */
#define spi1 ((spi_inst_t * const)spi1_hw)

#if !defined(PICO_DEFAULT_SPI_INSTANCE) && defined(PICO_DEFAULT_SPI)
#define PICO_DEFAULT_SPI_INSTANCE (__CONCAT(spi,PICO_DEFAULT_SPI))
#endif

#ifdef PICO_DEFAULT_SPI_INSTANCE
#define spi_default PICO_DEFAULT_SPI_INSTANCE
#endif

/*! \brief Convert SPI instance to hardware instance number
 *  \ingroup hardware_spi
 *
 * \param spi SPI instance
 * \return Number of SPI, 0 or 1.
 */
static inline uint spi_get_index(const spi_inst_t *spi) {
    invalid_params_if(SPI, spi != spi0 && spi != spi1);
    return spi == spi1 ? 1 : 0;
}

static inline spi_hw_t *spi_get_hw(spi_inst_t *spi) {
    spi_get_index(spi); // check it is a hw spi
    return (spi_hw_t *)spi;
}

static inline const spi_hw_t *spi_get_const_hw(const spi_inst_t *spi) {
    spi_get_index(spi);  // check it is a hw spi
    return (const spi_hw_t *)spi;
}
/*! \brief Write/Read to/from an SPI device
 *  \ingroup hardware_spi
 *
 * Write \p len bytes from \p src to SPI. Simultaneously read \p len bytes from SPI to \p dst.
 * Blocks until all data is transferred. No timeout, as SPI hardware always transfers at a known data rate.
 *
 * \param spi SPI instance specifier, either \ref spi0 or \ref spi1
 * \param src Buffer of data to write
 * \param dst Buffer for read data
 * \param len Length of BOTH buffers
 * \return Number of bytes written/read
*/
int spi_write_read_blocking(spi_inst_t *spi, const uint8_t *src, uint8_t *dst, size_t len);

/*! \brief Write to an SPI device, blocking
 *  \ingroup hardware_spi
 *
 * Write \p len bytes from \p src to SPI, and discard any data received back
 * Blocks until all data is transferred. No timeout, as SPI hardware always transfers at a known data rate.
 *
 * \param spi SPI instance specifier, either \ref spi0 or \ref spi1
 * \param src Buffer of data to write
 * \param len Length of \p src
 * \return Number of bytes written/read
 */
int spi_write_blocking(spi_inst_t *spi, const uint8_t *src, size_t len);

/*! \brief Read from an SPI device
 *  \ingroup hardware_spi
 *
 * Read \p len bytes from SPI to \p dst.
 * Blocks until all data is transferred. No timeout, as SPI hardware always transfers at a known data rate.
 * \p repeated_tx_data is output repeatedly on TX as data is read in from RX.
 * Generally this can be 0, but some devices require a specific value here,
 * e.g. SD cards expect 0xff
 *
 * \param spi SPI instance specifier, either \ref spi0 or \ref spi1
 * \param repeated_tx_data Buffer of data to write
 * \param dst Buffer for read data
 * \param len Length of buffer \p dst
 * \return Number of bytes written/read
 */
int spi_read_blocking(spi_inst_t *spi, uint8_t repeated_tx_data, uint8_t *dst, size_t len);

#ifdef __cplusplus
}
#endif

#endif