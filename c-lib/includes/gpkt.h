// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

#ifndef GPKT_H_
#define GPKT_H_

#include <pthread.h>

#ifdef __cplusplus
extern "C" {
#endif

#define GPKT_API __attribute__((visibility("default")))

#include <tlog.h>
#include <dpdk_api.h>
#include <msgchan.h>
#include <bits/pthreadtypes.h>

enum {
    DEFAULT_MSGCHAN_SIZE              = 1024,
    DEFAULT_MBUFS_PER_PORT_MULTIPLIER = 2,
};

#define MAX_MBUFS_PER_PORT(rxd, txd) ((rxd + txd) * DEFAULT_MBUFS_PER_PORT_MULTIPLIER)

/*
 * Some NICs require >= 2KB buffer as a receive data buffer. DPDK uses 2KB + HEADROOM (128) as
 * the default MBUF buffer size. This would make the pktmbuf buffer 2KB + HEADROOM +
 * sizeof(rte_mbuf) which is 2048 + 128 + 128 = 2304 mempool buffer size.
 *
 * For Jumbo frame buffers lets use MTU 9216 + FCS(4) + L2(14) = 9234, for buffer size we use 10KB
 */
#define _MBUF_LEN (PG_JUMBO_FRAME_LEN + RTE_PKTMBUF_HEADROOM + sizeof(struct rte_mbuf))

// Go-Pktgen global variables
typedef struct gpkt_s {
    pthread_t thread;                                                // Thread ID for DPDK thread
    msgchan_t *dpdk_chnl;                                            // Message channel for DPDK
    volatile int quit[RTE_MAX_LCORE];                                // Flag to quit the DPDK thread
    logical_core_t lcores[RTE_MAX_LCORE] __rte_cache_aligned;        // Logical cores
    physical_port_t ports[RTE_MAX_ETHPORTS] __rte_cache_aligned;        // Physical ports
} gpkt_t __rte_cache_aligned;

extern gpkt_t *gpkt;

#define ARGV_MAX_NUM  64         // Maximum number of command-line arguments
#define ARGV_MAX_SIZE 128        // Maximum size of each command-line argument

typedef struct gpkt_args_s {
    int argc;                                          // Number of command-line arguments
    char *argv[ARGV_MAX_NUM];                          // Command-line arguments
    char argv_str[ARGV_MAX_NUM][ARGV_MAX_SIZE];        // Concatenated command-line arguments
    pthread_barrier_t barrier;                         // Barrier for synchronization
} gpkt_args_t;

#ifdef __cplusplus
}
#endif

#endif /* GPKT_H_ */
