// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2023-2024 Intel Corporation

#include <stdio.h>
#include <unistd.h>

#include <rte_malloc.h>

#include <tlog.h>
#include <port.h>

enum {
    RX_PTHRESH = 8, /**< Default values of RX prefetch threshold reg. */
    RX_HTHRESH = 8, /**< Default values of RX host threshold reg. */
    RX_WTHRESH = 4, /**< Default values of RX write-back threshold reg. */

    TX_PTHRESH     = 36, /**< Default values of TX prefetch threshold reg. */
    TX_HTHRESH     = 0,  /**< Default values of TX host threshold reg. */
    TX_WTHRESH     = 0,  /**< Default values of TX write-back threshold reg. */
    TX_WTHRESH_1GB = 16, /**< Default value for 1GB ports */
};

static struct rte_eth_conf default_port_conf = {
    .rxmode =
        {
            .mq_mode          = RTE_ETH_MQ_RX_RSS,
            .max_lro_pkt_size = RTE_ETHER_MAX_LEN,
            .offloads         = RTE_ETH_RX_OFFLOAD_CHECKSUM,
        },

    .rx_adv_conf =
        {
            .rss_conf =
                {
                    .rss_key = NULL,
                    .rss_hf  = RTE_ETH_RSS_IP | RTE_ETH_RSS_TCP | RTE_ETH_RSS_UDP |
                              RTE_ETH_RSS_SCTP | RTE_ETH_RSS_L2_PAYLOAD,
                },
        },
    .txmode =
        {
            .mq_mode = RTE_ETH_MQ_TX_NONE,
        },
    .intr_conf =
        {
            .lsc = 0,
        },
};

#if 0
static int
config_ports(void)
{
    struct rte_eth_conf conf;
    uint16_t pid, nb_ports;
    port_info_t *pinfo;
    int32_t ret, sid;
    uint16_t nb_rxd = DEFAULT_RX_DESC, nb_txd = DEFAULT_TX_DESC, port;

    /* Find out the total number of ports in the system. */
    nb_ports = rte_eth_dev_count_avail();
    if (nb_ports == 0) {
        tlog_printf("*** Did not find any ports to use ***\n");
        return -1;
    }
    if (nb_ports > RTE_MAX_ETHPORTS) {
        tlog_printf("*** Too many ports in the system %d ***\n", nb_ports);
        return -1;
    }

    /* For each lcore setup each port that is handled by that lcore. */
    for (uint16_t lid = 0; lid < RTE_MAX_LCORE; lid++) {
        if ((pid = l2p_get_pid_by_lcore(lid)) >= RTE_MAX_ETHPORTS)
            continue;

        sid = rte_eth_dev_socket_id(pid);

        pinfo = l2p_get_port_pinfo(pid);
        if (pinfo == NULL) {
            /* Allocate each port_info_t structure on the correct NUMA node for the port */
            pinfo = rte_zmalloc_socket(NULL, sizeof(port_info_t), RTE_CACHE_LINE_SIZE, sid);
            if (!pinfo)
                rte_exit(EXIT_FAILURE, "Cannot allocate memory for port_info_t\n");

            pinfo->pid = pid;

            pinfo->fill_pattern_type = ABC_FILL_PATTERN;
            snprintf(pinfo->user_pattern, sizeof(pinfo->user_pattern), "%s", "0123456789abcdef");

            size_t pktsz = RTE_ETHER_MAX_LEN;

            l2p_set_port_pinfo(pid, pinfo);
        }
    }

    RTE_ETH_FOREACH_DEV(pid)
    {
        pinfo = l2p_get_port_pinfo(pid);
        port  = l2p_get_port(pid);
        if (pinfo == NULL || port == NULL)
            continue;

        /* grab the socket id value based on the pid being used. */
        sid = rte_eth_dev_socket_id(pid);

        rte_eth_dev_info_get(pid, &pinfo->dev_info);

        tlog_printf("Initialize Port %u ...\n", pid);

        /* Get a clean copy of the configuration structure */
        rte_memcpy(&conf, &default_port_conf, sizeof(struct rte_eth_conf));

        // if (pktgen.flags & JUMBO_PKTS_FLAG) {
        //     conf.rxmode.max_lro_pkt_size = RTE_ETHER_MAX_JUMBO_FRAME_LEN;
        //     if (pinfo->dev_info.tx_offload_capa & RTE_ETH_TX_OFFLOAD_MULTI_SEGS)
        //         conf.txmode.offloads |= RTE_ETH_TX_OFFLOAD_MULTI_SEGS;
        // }

        conf.rx_adv_conf.rss_conf.rss_key = NULL;
        conf.rx_adv_conf.rss_conf.rss_hf &= pinfo->dev_info.flow_type_rss_offloads;
        if (pinfo->dev_info.max_rx_queues == 1)
            conf.rxmode.mq_mode = RTE_ETH_MQ_RX_NONE;

        if (pinfo->dev_info.max_vfs) {
            if (conf.rx_adv_conf.rss_conf.rss_hf != 0)
                conf.rxmode.mq_mode = RTE_ETH_MQ_RX_VMDQ_RSS;
        }

        pinfo->lsc_enabled = 0;
        if (*pinfo->dev_info.dev_flags & RTE_ETH_DEV_INTR_LSC) {
            conf.intr_conf.lsc = 1;
            pinfo->lsc_enabled = 1;
        }

        conf.rxmode.offloads &= pinfo->dev_info.rx_offload_capa;

        if ((ret = rte_eth_dev_configure(pid, l2p_get_rxcnt(pid), l2p_get_txcnt(pid), &conf)) < 0)
            tlog_printf("Cannot configure device: port=%d, Num queues %d,%d\n", pid,
                        l2p_get_rxcnt(pid), l2p_get_txcnt(pid));

        // ret = rte_eth_dev_adjust_nb_rx_tx_desc(pid, &pktgen.nb_rxd, &pktgen.nb_txd);
        // if (ret < 0)
        //     rte_exit(EXIT_FAILURE, "Can't adjust number of descriptors: port=%u:%s\n", pid,
        //              rte_strerror(-ret));

        if ((ret = rte_eth_macaddr_get(pid, &pinfo->eth_src_addr)) < 0)
            rte_exit(EXIT_FAILURE, "Can't get MAC address: err=%d, port=%u\n", ret, pid);

        ret = rte_eth_dev_set_ptypes(pid, RTE_PTYPE_UNKNOWN, NULL, 0);
        if (ret < 0)
            rte_exit(EXIT_FAILURE, "Port %u, Failed to disable Ptype parsing\n", pid);

        for (int q = 0; q < l2p_get_rxcnt(pid); q++) {
            struct rte_eth_rxconf rxq_conf;
            struct rte_eth_conf conf = {0};

            rte_eth_dev_conf_get(pid, &conf);

            rxq_conf          = pinfo->dev_info.default_rxconf;
            rxq_conf.offloads = conf.rxmode.offloads;

            ret = rte_eth_rx_queue_setup(pid, q, nb_rxd, sid, &rxq_conf, port->rx_mp);
            if (ret < 0)
                tlog_printf("rte_eth_rx_queue_setup: err=%d, port=%d, %s\n", ret, pid,
                            rte_strerror(-ret));
        }

        for (int q = 0; q < l2p_get_txcnt(pid); q++) {
            struct rte_eth_txconf *txconf;

            txconf           = &pinfo->dev_info.default_txconf;
            txconf->offloads = default_port_conf.txmode.offloads;

            ret = rte_eth_tx_queue_setup(pid, q, nb_txd, sid, txconf);
            if (ret < 0)
                tlog_printf("rte_eth_tx_queue_setup: err=%d, port=%d, %s\n", ret, pid,
                            rte_strerror(-ret));
        }

        if (rte_eth_promiscuous_enable(pid))
            rte_exit(EXIT_FAILURE, "Enabling promiscuous failed: %s\n", rte_strerror(-rte_errno));

        /* Start device */
        if ((ret = rte_eth_dev_start(pid)) < 0)
            tlog_printf("rte_eth_dev_start: port=%d, %s\n", pid, rte_strerror(-ret));
    }
}
#endif

// Initialize ports
int
init_ports(void)
{
    tlog_printf("%s: started\n", __func__);

    // Implement port initialization logic here
    (void)default_port_conf;

#if 0
    if (config_ports() < 0) {
        tlog_printf("%s: failed to configure ports\n", __func__);
        return -1;
    }
#endif
    return 0;
}
