// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

#ifndef DPDK_API_H_
#define DPDK_API_H_

#include <stdint.h>
#include <stdbool.h>

#include <pthread.h>

#include <gpkt.h>
#include <msgchan.h>

#ifdef __cplusplus
extern "C" {
#endif

#define MODE_STRINGS                                   \
    {                                                  \
        "Unknown", "Main", "RxOnly", "TxOnly", "Rx/Tx" \
    }
enum { UNKNOWN_MODE = 0, MAIN_MODE = 1, RXONLY_MODE = 2, TXONLY_MODE = 3, RXTX_MODE = 4 };

// This matches the internal/gpcommon.go PhyiscalPort structure
typedef struct physical_port_s {
    uint16_t port_id;              // Port ID
    uint16_t num_rx_queues;        // Number of receive queues
    uint16_t num_tx_queues;        // Number of transmit queues
} physical_port_t;

// This matches the internal/gpcommon.go LogicalPort structure
typedef struct logical_port_s {
    physical_port_t *physical_port;        // Physical port information
    uint32_t lport_id;                     // Logical port ID
    uint16_t rx_qid;                       // Receive queue ID
    uint16_t tx_qid;                       // Transmit queue ID
} logical_port_t;

// This matches the internal/gpcommon.go CoreInfo structure
typedef struct logical_core_s {
    logical_port_t *logical_port;        // Logical port information
    uint16_t mode;                       // Mode (0: Receive, 1: Transmit)
    uint16_t core_id;                    // Core ID
} logical_core_t;

// This matches the L2pConfig structure
typedef struct l2p_config_s {
    uint32_t lport_id;             // Lport ID
    uint16_t core_id;              // Core ID
    uint16_t mode;                 // Mode (0: Receive, 1: Transmit)
    uint16_t rx_qid;               // Rx Queue ID
    uint16_t tx_qid;               // Tx Queue ID
    uint16_t port_id;              // Port ID
    uint16_t num_rx_queues;        // Number of receive queues
    uint16_t num_tx_queues;        // Number of transmit queues
    uint16_t reserved;             // Reserved
} l2p_config_t;

/**
 * @brief Starts DPDK.
 *
 * This function initializes the gpktApi library and performs any necessary setup.
 * It should be called before any other gpktApi functions are used.
 *
 * @param log_path The path to the log file.
 *
 * @return 0 on success, or a -1 on error.
 */
GPKT_API int dpdk_startup(char *log_path);

/**
 * @brief Stops DPDK.
 *
 * @returns N/A
 */
GPKT_API void dpdk_shutdown(void);

/**
 * @brief Adds an argument to the DPDK argv list.
 *
 * @param arg The argument to be added.
 * @return 0 on success, or a -1 on error when too many arguments are provided.
 */
GPKT_API int dpdk_add_argv(char *arg);

/**
 * @brief Launches a function in a separate thread.
 *
 * This function creates a new thread and executes the provided function in it.
 * The function is expected to be a DPDK application function that takes a void pointer
 * as an argument and returns an integer.
 *
 * @param arg The argument to be passed to the function. This parameter is marked as
 *            unused to avoid compiler warnings.
 *
 * @return 0 on success, or a non-zero value on error. The specific error code can be
 *         obtained by calling rte_errno().
 *
 * @note This function is intended to be used with DPDK applications. It assumes that
 *       the DPDK environment is already set up and initialized.
 */
GPKT_API int launch_func(void *arg __rte_unused);

GPKT_API void dpdk_l2p_dump(uint16_t core_id);

GPKT_API int dpdk_l2p_config(l2p_config_t *cfg);

GPKT_API void dpdk_l2p_config_dump(l2p_config_t *cfg);

#ifdef __cplusplus
}
#endif

#endif /* DPDK_API_H_ */
