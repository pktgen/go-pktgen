/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation
 */

#ifndef _MSGCHAN_PRIV_H_
#define _MSGCHAN_PRIV_H_

#include <sys/queue.h>
#include <rte_common.h>
#include <rte_ring.h>
#include "msgchan.h"

/**
 * @file
 * Private Message Channels information
 *
 * Private data structures and information for msgchan library. The external msgchan pointer
 * is a void pointer and converted to the msg_chan_t structure pointer in the code.
 */

#ifdef __cplusplus
extern "C" {
#endif

#define MC_COOKIE ('C' << 24 | 'h' << 16 | 'a' << 8 | 'n')

typedef struct msg_chan {
    TAILQ_ENTRY(msg_chan) next;             // Next entry in the global list.
    TAILQ_HEAD(, msg_chan) children;        // List of attached children
    char name[RTE_RING_NAMESIZE];           // The name of the message channel
    uint32_t nchildren;                     // Number of children
    bool mutex_inited;                      // Flag to detect mutex is inited
    struct rte_ring *rings[2];              // Pointers to the send/recv rings
    struct msg_chan *parent;                // Pointer to parent channel if a child.
    pthread_mutex_t mutex;                  // Mutex to protect the attached list
    uint32_t cookie;                        // Cookie value to test for valid entry
    uint64_t send_calls;                    // Number of send calls
    uint64_t send_cnt;                      // Number of objects sent
    uint64_t recv_calls;                    // Number of receive calls
    uint64_t recv_cnt;                      // Number of objects received
    uint64_t recv_timeouts;                 // Number of receive timeouts
} msg_chan_t;

#ifdef __cplusplus
}
#endif

#endif /* _MSGCHAN_PRIV_H_ */
