// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

#include <stdio.h>
#include <unistd.h>

#include <rte_malloc.h>
#include <rte_ethdev.h>

#include <gpkt.h>
#include <port.h>

enum {
    RX_PTHRESH     = 8,  /**< Default values of RX prefetch threshold reg. */
    RX_HTHRESH     = 8,  /**< Default values of RX host threshold reg. */
    RX_WTHRESH     = 4,  /**< Default values of RX write-back threshold reg. */
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

static port_info_t *port_infos[RTE_MAX_ETHPORTS];

static int
port_alloc(port_config_t *cfg)
{
    size_t size     = sizeof(port_info_t);
    port_info_t *pi = NULL;
    int sid;

    if (cfg == NULL)
        TLOG_ERR_RET("Invalid port configuration\n");
    if (cfg->pid >= RTE_MAX_ETHPORTS)
        TLOG_ERR_RET("Invalid port_id %u\n", cfg->pid);

    if ((pi = port_info_get(cfg->pid)) != NULL)
        return 0;

    if ((sid = rte_eth_dev_socket_id(cfg->pid)) < 0)
        TLOG_ERR_RET("rte_eth_dev_socket_id() failed\n");

    TLOG_PRINT("Allocating port_info_t for port %u on socket %d, size %ld\n", cfg->pid, sid, size);

    /* Allocate each port_info_t structure on the correct NUMA node for the port */
    pi = rte_zmalloc_socket(NULL, size, RTE_CACHE_LINE_SIZE, sid);
    if (!pi)
        rte_exit(EXIT_FAILURE, "Cannot allocate memory for port_info_t\n");

    pi->fill_pattern_type = ABC_FILL_PATTERN;
    snprintf(pi->user_pattern, sizeof(pi->user_pattern), "%s", "0123456789abcdef");

    pi->pid               = cfg->pid;
    pi->sid               = sid;
    pi->rxqcnt            = cfg->rxqcnt;
    pi->txqcnt            = cfg->txqcnt;
    pi->rx_burst          = cfg->rx_burst;
    pi->tx_burst          = cfg->tx_burst;
    pi->nb_rxd            = cfg->nb_rxd;
    pi->nb_txd            = cfg->nb_txd;
    pi->nb_mbufs_per_port = cfg->nb_mbufs_per_port;
    pi->cache_size        = cfg->cache_size;

    pi->pkt = rte_zmalloc_socket(NULL, sizeof(pkt_t), RTE_CACHE_LINE_SIZE, sid);
    if (!pi->pkt)
        rte_exit(EXIT_FAILURE, "Cannot allocate memory for packet data\n");

    pi->pkt->tcp_flags = DEFAULT_TCP_FLAGS;
    pi->pkt->tcp_seq   = DEFAULT_TCP_SEQ_NUMBER;
    pi->pkt->tcp_ack   = DEFAULT_TCP_ACK_NUMBER;

    if (rte_eth_dev_info_get(pi->pid, &pi->dev_info) < 0)
        rte_exit(EXIT_FAILURE, "unable to get device information\n");

    port_infos[pi->pid] = pi;

    return 0;
}

int
port_set_info(port_config_t *cfg)
{
    TLOG_PRINT("Setting port %d info: rxcnt %u, txcnt %u\n", cfg->pid, cfg->rxqcnt, cfg->txqcnt);

    if (port_alloc(cfg) < 0)
        TLOG_ERR_RET("unable to allocate port information structure %d\n", cfg->pid);

    return 0;
}

port_config_t *
port_get_info(uint16_t port_id)
{
    port_config_t *cfg = calloc(1, sizeof(port_config_t));
    port_info_t *pi;

    if (cfg == NULL)
        TLOG_NULL_RET("failed to allocate port_config_t structure\n");

    if (port_id >= RTE_MAX_ETHPORTS)
        TLOG_NULL_RET("port %d invalid\n", port_id);

    pi = port_infos[port_id];

    if (pi == NULL)
        TLOG_NULL_RET("port %d not found\n", port_id);

    cfg->pid               = pi->pid;
    cfg->rxqcnt            = pi->rxqcnt;
    cfg->txqcnt            = pi->txqcnt;
    cfg->nb_rxd            = pi->nb_rxd;
    cfg->nb_txd            = pi->nb_txd;
    cfg->rx_burst          = pi->rx_burst;
    cfg->tx_burst          = pi->tx_burst;
    cfg->cache_size        = pi->cache_size;
    cfg->nb_mbufs_per_port = pi->nb_mbufs_per_port;

    return cfg;
}

void
port_free_info(port_config_t *cfg)
{
    free(cfg);
}

port_info_t *
port_info_get(uint16_t port_id)
{
    if (port_id >= RTE_MAX_ETHPORTS)
        TLOG_NULL_RET("port %d invalid\n", port_id);

    return port_infos[port_id];
}

static struct rte_mempool *
create_pktmbuf_pool(const char *type, port_info_t *pi)
{
    char name[RTE_MEMZONE_NAMESIZE];

    /* Create the pktmbuf pool one per lcore/port */
    snprintf(name, sizeof(name), "%s-%u", type, pi->pid);
    return rte_pktmbuf_pool_create(name, pi->nb_mbufs_per_port, pi->cache_size, 0,
                                   RTE_MBUF_DEFAULT_BUF_SIZE, pi->sid);
}

static int
port_setup(uint16_t port_id)
{
    port_info_t *pi;
    struct rte_eth_conf conf;
    int ret;

    TLOG_PRINT(">>> Setting up port %u on core %d\n", port_id, rte_lcore_id());

    pi = port_info_get(port_id);
    if (pi == NULL)
        TLOG_ERR_RET("unable to allocate port information structure %d\n", port_id);

    TLOG_PRINT("Initializing port %u == %u\n", port_id, pi->pid);

    // Create a mempool one per port/queue.
    pi->rx_mp = NULL;
    if (pi->rx_mp == NULL) {
        if ((pi->rx_mp = create_pktmbuf_pool("Rx", pi)) == NULL)
            rte_panic("Cannot create Rx mbuf pool for %d\n", pi->nb_mbufs_per_port);
    }

    /* Get a clean copy of the configuration structure */
    rte_memcpy(&conf, &default_port_conf, sizeof(struct rte_eth_conf));

    if (pi->flags & JUMBO_PKTS_FLAG) {
        conf.rxmode.max_lro_pkt_size = RTE_ETHER_MAX_JUMBO_FRAME_LEN;
        if (pi->dev_info.tx_offload_capa & RTE_ETH_TX_OFFLOAD_MULTI_SEGS)
            conf.txmode.offloads |= RTE_ETH_TX_OFFLOAD_MULTI_SEGS;
    }

    conf.rx_adv_conf.rss_conf.rss_key = NULL;
    conf.rx_adv_conf.rss_conf.rss_hf &= pi->dev_info.flow_type_rss_offloads;
    if (pi->dev_info.max_rx_queues == 1)
        conf.rxmode.mq_mode = RTE_ETH_MQ_RX_NONE;

    if (pi->dev_info.max_vfs) {
        if (conf.rx_adv_conf.rss_conf.rss_hf != 0)
            conf.rxmode.mq_mode = RTE_ETH_MQ_RX_VMDQ_RSS;
    }

    pi->lsc_enabled = 0;
    if (*pi->dev_info.dev_flags & RTE_ETH_DEV_INTR_LSC) {
        conf.intr_conf.lsc = 1;
        pi->lsc_enabled    = 1;
    }

    conf.rxmode.offloads &= pi->dev_info.rx_offload_capa;

    if ((ret = rte_eth_dev_configure(port_id, pi->rxqcnt, pi->txqcnt, &conf)) < 0)
        rte_panic("Cannot configure device: port=%d, Num queues %d,%d", port_id, pi->rxqcnt,
                  pi->txqcnt);

    ret = rte_eth_dev_adjust_nb_rx_tx_desc(port_id, &pi->nb_rxd, &pi->nb_txd);
    if (ret < 0)
        rte_panic("Can't adjust number of descriptors: port=%u:%s\n", port_id, rte_strerror(-ret));

    TLOG_PRINT("Port %u: Number Rx/Tx descriptors %u/%u\n", port_id, pi->nb_rxd, pi->nb_txd);

    if ((ret = rte_eth_macaddr_get(port_id, &pi->eth_src_addr)) < 0)
        rte_panic("Can't get MAC address: err=%d, port=%u\n", ret, port_id);
    rte_ether_addr_copy(&pi->eth_src_addr, &pi->pkt->eth_src_addr);

    if ((ret = rte_eth_dev_set_ptypes(port_id, RTE_PTYPE_UNKNOWN, NULL, 0)) < 0)
        rte_exit(EXIT_FAILURE, "Port %u, Failed to disable Ptype parsing\n", port_id);

    TLOG_PRINT("Port %u: MAC address: %02x:%02x:%02x:%02x:%02x:%02x rxcnt %d\n", port_id,
               pi->eth_src_addr.addr_bytes[0], pi->eth_src_addr.addr_bytes[1],
               pi->eth_src_addr.addr_bytes[2], pi->eth_src_addr.addr_bytes[3],
               pi->eth_src_addr.addr_bytes[4], pi->eth_src_addr.addr_bytes[5], pi->rxqcnt);

    for (int q = 0; q < pi->rxqcnt; q++) {
        struct rte_eth_rxconf rxq_conf;
        struct rte_eth_conf conf = {0};

        if (rte_eth_dev_conf_get(port_id, &conf) < 0)
            rte_panic("rte_eth_dev_conf_get: err=%d, port=%d\n", ret, port_id);

        rxq_conf          = pi->dev_info.default_rxconf;
        rxq_conf.offloads = conf.rxmode.offloads;

        TLOG_PRINT("Rx setup Port %u, Queue %u\n", port_id, q);
        ret = rte_eth_rx_queue_setup(port_id, q, pi->nb_rxd, pi->sid, &rxq_conf, pi->rx_mp);
        if (ret < 0)
            rte_panic("rte_eth_rx_queue_setup: err=%d, port=%d, %s", ret, port_id,
                      rte_strerror(-ret));
    }
    TLOG_PRINT("Port %u: Number of RX queues %u\n", port_id, pi->rxqcnt);

    for (int q = 0; q < pi->txqcnt; q++) {
        struct rte_eth_txconf *txconf;

        txconf           = &pi->dev_info.default_txconf;
        txconf->offloads = default_port_conf.txmode.offloads;

        TLOG_PRINT("Tx setup Port %u, Queue %u\n", port_id, q);
        if ((ret = rte_eth_tx_queue_setup(port_id, q, pi->nb_txd, pi->sid, txconf)) < 0)
            rte_panic("rte_eth_tx_queue_setup: err=%d, port=%d, %s", ret, port_id,
                      rte_strerror(-ret));
    }
    TLOG_PRINT("Port %u: Number of TX queues %u\n", port_id, pi->txqcnt);

    if (rte_eth_promiscuous_enable(port_id))
        rte_panic("Enabling promiscuous failed: %s\n", rte_strerror(-rte_errno));

    pi->pkt->pkt_size = RTE_ETHER_MIN_LEN - RTE_ETHER_CRC_LEN;

    /* Start device */
    if ((ret = rte_eth_dev_start(port_id)) < 0)
        rte_panic("port=%d, %s", port_id, rte_strerror(-ret));

    TLOG_PRINT("Port %u, Device started\n", port_id);
    return 0;
}

// Initialize ports
int
port_init(uint16_t pid)
{
    uint16_t nb_ports;

    /* Find out the number of ports available in the system. */
    nb_ports = rte_eth_dev_count_avail();
    if (nb_ports == 0 || nb_ports > RTE_MAX_ETHPORTS || pid >= nb_ports)
        TLOG_ERR_RET("*** Did not find any ports to use or too many %u ***\n", nb_ports);

    // Setup the port and configure DPDK resources
    return port_setup(pid);
}

int
port_ether_stats(uint16_t pid, void *stats_ptr)
{
    struct rte_eth_stats *stats = (struct rte_eth_stats *)stats_ptr;

    if (rte_eth_stats_get(pid, stats) < 0)
        TLOG_ERR_RET("rte_eth_stats_get: err=%d, port=%u\n", rte_errno, pid);

    return 0;
}

int
port_packet_stats(uint16_t pid, void *stats_ptr)
{
    pkt_stats_t *stats = (pkt_stats_t *)stats_ptr;
    port_info_t *pi    = port_info_get(pid);

    if (pi == NULL)
        TLOG_ERR_RET("Port info not found for port %u\n", pid);

    *stats = pi->pkt_stats;

    return 0;
}

uint64_t
port_link_status(uint16_t pid)
{
    struct rte_eth_link link;
    int ret;

    if ((ret = rte_eth_link_get(pid, &link)) < 0)
        TLOG_ERR_RET("rte_eth_link_get: err=%d, port=%u\n", ret, pid);

    return link.val64;
}

int
port_mac_address(uint16_t port_id, void *mac)
{
    struct rte_ether_addr *eth_addr = (struct rte_ether_addr *)mac;
    int ret;

    if ((ret = rte_eth_macaddr_get(port_id, eth_addr)) < 0)
        TLOG_ERR_RET("Can't get MAC address: err=%d, port=%u\n", ret, port_id);

    return 0;
}

int
port_device_info(uint16_t port_id, void *dev_info)
{
    struct rte_eth_dev_info dev = {0};
    device_info_t *info         = (device_info_t *)dev_info;
    struct rte_bus *bus;
    int ret;

    if ((ret = rte_eth_dev_info_get(port_id, &dev)) < 0)
        TLOG_ERR_RET("rte_eth_dev_info_get: err=%d, port=%u\n", ret, port_id);

    strncpy(info->name, rte_dev_name((const struct rte_device *)dev.device),
            sizeof(info->name) - 1);

    bus = rte_bus_find_by_device((const struct rte_device *)dev.device);
    if (bus)
        strncpy(info->bus_name, rte_bus_name(bus), sizeof(info->bus_name) - 1);
    else
        strncpy(info->bus_name, "Unknown", sizeof(info->bus_name));

    if ((ret = rte_eth_macaddr_get(port_id, &info->mac_addr)) < 0)
        TLOG_ERR_RET("Can't get MAC address: err=%d, port=%u\n", ret, port_id);

    info->if_index           = dev.if_index;
    info->min_mtu            = dev.min_mtu;
    info->max_mtu            = dev.max_mtu;
    info->min_rx_bufsize     = dev.min_rx_bufsize;
    info->max_rx_bufsize     = dev.max_rx_bufsize;
    info->max_rx_pktlen      = dev.max_rx_pktlen;
    info->max_rx_queues      = dev.max_rx_queues;
    info->max_tx_queues      = dev.max_tx_queues;
    info->max_mac_addrs      = dev.max_mac_addrs;
    info->max_hash_mac_addrs = dev.max_hash_mac_addrs;
    info->max_vfs            = dev.max_vfs;
    info->nb_rx_queues       = dev.nb_rx_queues;
    info->nb_tx_queues       = dev.nb_tx_queues;
    info->socket_id          = rte_eth_dev_socket_id(port_id);

    return 0;
}
