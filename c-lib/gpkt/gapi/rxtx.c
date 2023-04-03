// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

#include <stdio.h>
#include <unistd.h>

#include <rte_malloc.h>
#include <rte_ethdev.h>

#include <gpkt.h>
#include <port.h>

static void
port_main_receive(uint16_t pid, uint16_t qid, struct rte_mbuf **pkts_burst, uint16_t nb_pkts)
{
    uint16_t nb_rx;

    /* Read packets from RX queues and free the mbufs */
    if (likely((nb_rx = rte_eth_rx_burst(pid, qid, pkts_burst, nb_pkts)) > 0))
        rte_pktmbuf_free_bulk(pkts_burst, nb_rx);
}

void *
port_rxtx_loop(gpkt_t *g, uint16_t pid, uint16_t rx_qid, uint16_t tx_qid)
{
    port_info_t *p    = port_get(pid);
    uint16_t rx_burst = p->rx_burst;
    uint16_t tx_burst = p->tx_burst;
    struct rte_mbuf *pkts_burst[rx_burst];
    int lid = rte_lcore_id();

    (void)tx_qid;

    port_init(pid);

    TLOG_PRINT("Starting RX/TX loop on %d core, port %u, Rx/Tx queues %u/%u, burst %u/%u\n", lid,
               pid, rx_qid, tx_qid, rx_burst, tx_burst);

    while (!g->quit[lid]) {
        port_main_receive(pid, rx_qid, pkts_burst, rx_burst);
    }

    return NULL;
}
