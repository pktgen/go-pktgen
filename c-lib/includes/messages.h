// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

#ifndef _GPKT_MESSAGES_H_
#define _GPKT_MESSAGES_H_

#include <gpkt.h>
#include <msgchan.h>
#include "tlog.h"

#ifdef __cplusplus
extern "C" {
#endif

enum { UNKNOWN_MSG, EXIT_MSG, LAUNCH_MSG, PORT_MSG, MAX_MSGS };

#define MSG_STRINGS                                  \
    {                                                \
        "NOOP", "EXIT", "LAUNCH", "PORT", "Unknown", \
    }

typedef struct {
    uint32_t reserved1;        // Reserved for future use
    uint32_t reserved2;        // Reserved for future use
} exit_msg_t;

typedef struct {
    uint32_t call_main;        // Skip MAIN lcore, CALL_MAIN = 1, SKIP_MAIN = 0
} launch_msg_t;

typedef struct {
    uint32_t portlist;        // Port list for start_msg
    uint32_t enable;          // Start or stop port(s), 0 = stop, 1 = start
} start_stop_msg_t;

typedef int (*msg_func_t)(gpkt_t *g, mc_msg_t *msg);
GPKT_API int msg_channel_process(msgchan_t *mc);

#ifdef __cplusplus
}
#endif

#endif /* _GPKT_MESSAGES_H_ */
