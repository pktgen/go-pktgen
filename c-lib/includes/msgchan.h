/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

#ifndef _MSGCHAN_H_
#define _MSGCHAN_H_

#include <sys/queue.h>
#include <rte_common.h>
#include <rte_ring.h>

#include <gpkt.h>

/**
 * @file
 * Message Channels
 *
 * Create a message channel using two lockless rings to communicate between two threads.
 *
 * Message channels are similar to pipes in Linux and other platforms, but does not support
 * message passing between processes.
 *
 */

#ifdef __cplusplus
extern "C" {
#endif

#define MC_RECV_RING 0        // Receive index into msgchan_t.rings
#define MC_SEND_RING 1        // Send index into msgchan_t.rings

typedef struct mc_info {
    struct rte_ring *recv_ring;        // Pointers to the recv ring
    struct rte_ring *send_ring;        // Pointers to the send ring
    uint64_t send_calls;               // Number of send calls
    uint64_t send_cnt;                 // Number of objects sent
    uint64_t recv_calls;               // Number of receive calls
    uint64_t recv_cnt;                 // Number of objects received
    uint64_t recv_timeouts;            // Number of receive timeouts
    int child_count;                   // Number of children
} mc_info_t;

typedef void msgchan_t;        // Opaque msgchan structure pointer

typedef struct mc_msg_s {
    uint16_t action;          // Action to be performed gpkt_msg_e values
    uint16_t len;             // Length of the message in bytes
    uint32_t reserved;        // Reserved for future use
    uint64_t data[RTE_CACHE_LINE_SIZE / sizeof(uint64_t) - 1];        // Message data 64-8 bytes
} mc_msg_t;

/**
 * @brief Create a message channel
 *
 * @param name
 *   The name of the message channel
 * @param sz
 *   The number of entries the message channel.
 * @return
 *   The pointer to the msgchan structure or NULL on error
 */
GPKT_API msgchan_t *mc_create(const char *name, uint32_t sz);

/**
 * @brief Attach to an existing message channel as a child.
 *
 * This function allows a process to attach to an existing message channel as a child.
 *
 * @param parent_name
 *   The name of the parent message channel to attach to.
 *
 * @return
 *   - A pointer to the msgchan structure if the attachment is successful.
 *   - NULL if the parent message channel does not exist
 */
GPKT_API msgchan_t *mc_attach(const char *parent_name);

/**
 * @brief Destroy the message channel and free resources.
 *
 * @param mc
 *   The msgchan structure pointer to destroy
 * @return
 *   N/A
 */
GPKT_API void mc_destroy(msgchan_t *mc);

/**
 * @brief Send object messages to the other end of the channel
 *
 * @param mc
 *   The message channel structure pointer
 * @param objs
 *   An array of void *objects to send
 * @param count
 *   The number of entries in the objs array.
 * @return
 *   -1 on error or number of objects sent.
 */
GPKT_API int mc_send(msgchan_t *mc, void **objs, int count);

/**
 * @brief Receive message routine from other end of the channel
 *
 * @param mc
 *   The message channel structure pointer
 * @param objs
 *   An array of objects pointers to place the received objects pointers
 * @param count
 *   The number of entries in the objs array.
 * @param msec
 *   Number of milliseconds to wait for data, if return without waiting.
 * @return
 *   -1 on error or number of objects
 */
GPKT_API int mc_recv(msgchan_t *mc, void **objs, int count, uint64_t msec);

/**
 * @brief Lookup a message channel by name - parent only lookup
 *
 * @param name
 *   The name of the message channel to find, which is for parent channels
 * @return
 *   NULL if not found, otherwise the message channel pointer
 */
GPKT_API msgchan_t *mc_lookup(const char *name);

/**
 * @brief Return the name string for the msgchan_t pointer
 *
 * @param mc
 *   The message channel structure pointer
 * @return
 *   NULL if invalid pointer or string to message channel name
 */
GPKT_API const char *mc_name(msgchan_t *mc);

/**
 * @brief Return size and free space in the Producer/Consumer rings.
 *
 * @param mc
 *   The message channel structure pointer
 * @param recv_free_cnt
 *   The pointer to place the receive free count, can be NULL.
 * @param send_free_cnt
 *   The pointer to place the send free count, can be NULL.
 * @return
 *   -1 on error or size of the massage channel rings.
 */
GPKT_API int mc_size(msgchan_t *mc, int *recv_free_cnt, int *send_free_cnt);

/**
 * Return the message channel information structure data
 *
 * @param _mc
 *   The message channel structure pointer
 * @param info
 *   The pointer to the mc_info_t structure
 * @return
 *   0 on success or -1 on error
 */
GPKT_API int mc_info(msgchan_t *_mc, mc_info_t *info);

/**
 * @brief Dump out the details of the given message channel structure
 *
 * @param mc
 *   The message channel structure pointer
 * @return
 *   -1 if mc is NULL or 0 on success
 */
GPKT_API void mc_dump(msgchan_t *mc);

/**
 * @brief List out all message channels currently created.
 */
GPKT_API void mc_list(void);

#ifdef __cplusplus
}
#endif

#endif /* _MSGCHAN_H_ */
