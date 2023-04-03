// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

#include <stdio.h>
#include <unistd.h>

#include <rte_malloc.h>
#include <rte_ethdev.h>

#include <gpkt.h>
#include <port.h>

typedef enum {
    PACKET_CONSUMED = 0,
    UNKNOWN_PACKET  = 0xEEEE,
    DROP_PACKET     = 0xFFFE,
    FREE_PACKET     = 0xFFFF
} pktType_e;

/**
 *
 * pktgen_packet_type - Examine a packet and return the type of packet
 *
 * DESCRIPTION
 * Examine a packet and return the type of packet.
 * the packet.
 *
 * RETURNS: N/A
 *
 * SEE ALSO:
 */
static __inline__ pktType_e
pktgen_packet_type(struct rte_mbuf *m)
{
    pktType_e ret;
    struct rte_ether_hdr *eth;

    eth = rte_pktmbuf_mtod(m, struct rte_ether_hdr *);

    ret = ntohs(eth->ether_type);

    return ret;
}

/**
 *
 * pktgen_packet_classify - Examine a packet and classify it for statistics
 *
 * DESCRIPTION
 * Examine a packet and determine its type along with counting statistics around
 * the packet.
 *
 * RETURNS: N/A
 *
 * SEE ALSO:
 */
static void
packet_classify(struct rte_mbuf *m, int pid)
{
    port_info_t *pinfo     = port_info_get(pid);
    pkt_stats_t *pkt_stats = &pinfo->pkt_stats;
    uint16_t plen;
    pktType_e pType;

    pType = pktgen_packet_type(m);

    /* Count the type of packets found. */
    switch ((int)pType) {
    case RTE_ETHER_TYPE_ARP:
        pkt_stats->arp_pkts++;
        break;
    case RTE_ETHER_TYPE_IPV4:
        pkt_stats->ip_pkts++;
        break;
    case RTE_ETHER_TYPE_IPV6:
        pkt_stats->ipv6_pkts++;
        break;
    case RTE_ETHER_TYPE_VLAN:
        pkt_stats->vlan_pkts++;
        break;
    default:
        break;
    }

    plen = rte_pktmbuf_pkt_len(m) + RTE_ETHER_CRC_LEN;

    /* Count the size of each packet. */
    if (plen < RTE_ETHER_MIN_LEN)
        pkt_stats->runt++;
    else if (plen > RTE_ETHER_MAX_LEN)
        pkt_stats->jumbo++;
    else if (plen == RTE_ETHER_MIN_LEN)
        pkt_stats->_64++;
    else if ((plen >= (RTE_ETHER_MIN_LEN + 1)) && (plen <= 127))
        pkt_stats->_65_127++;
    else if ((plen >= 128) && (plen <= 255))
        pkt_stats->_128_255++;
    else if ((plen >= 256) && (plen <= 511))
        pkt_stats->_256_511++;
    else if ((plen >= 512) && (plen <= 1023))
        pkt_stats->_512_1023++;
    else if ((plen >= 1024) && (plen <= RTE_ETHER_MAX_LEN))
        pkt_stats->_1024_1518++;
    else {
        printf("Unknown packet size: %u", plen);
        pinfo->pkt_stats.unknown_pkts++;
    }

    uint8_t *p = rte_pktmbuf_mtod(m, uint8_t *);

    /* Process multicast and broadcast packets. */
    if (unlikely(p[0] & 1)) {
        if ((p[0] == 0xff) && (p[1] == 0xff))
            pkt_stats->broadcast++;
        else
            pkt_stats->multicast++;
    }
}

/**
 *
 * pktgen_packet_classify_bulk - Classify a set of packets in one call.
 *
 * DESCRIPTION
 * Classify a list of packets and to improve classify performance.
 *
 * RETURNS: N/A
 *
 * SEE ALSO:
 */
#define PREFETCH_OFFSET 3
static __inline__ void
packet_classify_bulk(int pid, struct rte_mbuf **pkts, int nb_rx)
{
    int j, i;

    /* Prefetch first packets */
    for (j = 0; j < PREFETCH_OFFSET && j < nb_rx; j++)
        rte_prefetch0(rte_pktmbuf_mtod(pkts[j], void *));

    /* Prefetch and handle already prefetched packets */
    for (i = 0; i < (nb_rx - PREFETCH_OFFSET); i++) {
        rte_prefetch0(rte_pktmbuf_mtod(pkts[j], void *));
        j++;

        packet_classify(pkts[i], pid);
    }

    /* Handle remaining prefetched packets */
    for (; i < nb_rx; i++)
        packet_classify(pkts[i], pid);
}

void *
port_rxtx_loop(gpkt_t *g, uint16_t pid, uint16_t rx_qid, uint16_t tx_qid)
{
    port_info_t *p    = port_info_get(pid);
    uint16_t rx_burst = p->rx_burst;
    uint16_t tx_burst = p->tx_burst;
    struct rte_mbuf *pkts_burst[rx_burst];
    int lid        = rte_lcore_id();
    uint16_t nb_rx = 0;

    (void)tx_qid;

    port_init(pid);

    TLOG_PRINT("Starting RX/TX loop on %d core, port %u, Rx/Tx queues %u/%u, burst %u/%u\n", lid,
               pid, rx_qid, tx_qid, rx_burst, tx_burst);

    while (!g->quit[lid]) {
        /* Read packets from RX queues and free the mbufs */
        if (likely((nb_rx = rte_eth_rx_burst(pid, rx_qid, pkts_burst, rx_burst)) > 0)) {
            packet_classify_bulk(pid, pkts_burst, nb_rx);
            rte_pktmbuf_free_bulk(pkts_burst, nb_rx);
        }
    }

    return NULL;
}
