// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

#include <stdio.h>
#include <unistd.h>

#include <rte_common.h>
#include <rte_launch.h>

#include <gpkt.h>
#include <port.h>
#include <single.h>
#include <pcap.h>
#include <messages.h>
#include <msgchan.h>

#define DPDK_CHANNEL_PREFIX "eal"

#define _(f) static int f(gpkt_t *g, mc_msg_t *msg);
_(process_unknown_msg)
_(process_exit_msg)
_(process_launch_msg)
_(process_port_msg)
#undef _

// clang-format off
static msg_func_t msg_func[MAX_MSGS] = {
#define _(u, f) [u] = f
    _(UNKNOWN_MSG, process_unknown_msg),
    _(EXIT_MSG,    process_exit_msg),
    _(LAUNCH_MSG,  process_launch_msg),
    _(PORT_MSG,    process_port_msg),
#undef _
};
// clang-format on

int
msg_channel_process(msgchan_t *mc)
{
    mc_msg_t msg              = {0};
    const char *msg_strings[] = MSG_STRINGS;

    // Get a message from the channel
    if (mc_recv(mc, (void **)&msg, 1, 0) == 0)
        rte_pause();
    else {
        uint16_t action = (msg.action >= MAX_MSGS) ? UNKNOWN_MSG : msg.action;

        if (msg.len > sizeof(msg.data))
            TLOG_ERR_RET("Message data exceeds buffer size, len %u\n", msg.len);

        TLOG_PRINT("Received {%s} message, len %u\n", msg_strings[action], msg.len);
        if (msg_func[action] == NULL || msg_func[action](mc, &msg) < 0)
            TLOG_ERR_RET("Error processing message %s, len %u\n", msg_strings[action], msg.len);
    }

    return 0;
}

static int
process_unknown_msg(gpkt_t *g __rte_unused, mc_msg_t *msg)
{
    TLOG_PRINT("Processing unknown message... (%d)\n", msg->action);
    (void)msg;

    return 0;
}

static int
process_exit_msg(gpkt_t *g, mc_msg_t *msg __rte_unused)
{
    uint32_t lcore_id;
    uint16_t port_id;

    TLOG_PRINT("Processing stop message...\n");

    g->quit[rte_lcore_id()] = 1;

    RTE_LCORE_FOREACH_WORKER(lcore_id)
    {
        if (rte_eal_wait_lcore(lcore_id) < 0)
            break;
    }

    RTE_ETH_FOREACH_DEV(port_id)
    {
        int ret;

        TLOG_PRINT("Closing port %d...", port_id);
        if ((ret = rte_eth_dev_stop(port_id)) != 0)
            TLOG_PRINT("rte_eth_dev_stop: err=%d, port=%d\n", ret, port_id);
        rte_eth_dev_close(port_id);
        TLOG_PRINT(" Done\n");
    }

    /* clean up the EAL */
    TLOG_PRINT("Cleaning up DPDK...\n");
    mc_destroy(g->dpdk_chnl);
    rte_eal_cleanup();

    TLOG_PRINT("DPDK Done\n");
    return 0;
}

static int
process_launch_msg(gpkt_t *g __rte_unused, mc_msg_t *msg)
{
    launch_msg_t *launch_msg = (launch_msg_t *)msg->data;

    TLOG_PRINT("Processing launch message...\n");

    // Launch each lcore unless main thread should be skipped
    if (rte_eal_mp_remote_launch(launch_func, g, launch_msg->call_main) < 0)
        TLOG_ERR_RET("Failed to launch Go-Pktgen thread\n");

    return 0;
}

static int
process_port_msg(gpkt_t *g __rte_unused, mc_msg_t *msg)
{
    start_stop_msg_t *port_msg = (start_stop_msg_t *)msg->data;

    if (port_msg->enable) {
        TLOG_PRINT("Processing start port, portlist %08x\n", port_msg->portlist);
    } else {
        TLOG_PRINT("Processing stop port, portlist %08x\n", port_msg->portlist);
    }

    return 0;
}
