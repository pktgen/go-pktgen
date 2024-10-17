// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2023-2024 Intel Corporation

#ifndef GPKT_PORT_H_
#define GPKT_PORT_H_

#include <stdint.h>

#include <rte_config.h>
#include <rte_config.h>
#include <rte_ether.h>
#include <rte_ethdev.h>
#include <rte_dev.h>

#if defined(RTE_LIBRTE_PMD_BOND) || defined(RTE_NET_BOND)
#include <rte_eth_bond_8023ad.h>
#endif
#include <rte_bus_pci.h>
#include <rte_bus.h>

#include <constants.h>
#include <stats.h>

typedef struct {
    uint32_t ip_addr;
    uint32_t netmask;
} ip_config_t;

#ifdef __cplusplus
extern "C" {
#endif

#define USER_PATTERN_SIZE 16
typedef enum {
    ZERO_FILL_PATTERN = 1,
    ABC_FILL_PATTERN,
    USER_FILL_PATTERN,
    NO_FILL_PATTERN,
} fill_t;

enum {
    URG_FLAG = 0x20,
    ACK_FLAG = 0x10,
    PSH_FLAG = 0x08,
    RST_FLAG = 0x04,
    SYN_FLAG = 0x02,
    FIN_FLAG = 0x01
};

enum {
    DEFAULT_NETMASK        = 0xFFFFFF00,
    DEFAULT_IP_ADDR        = (192 << 24) | (168 << 16),
    DEFAULT_TX_COUNT       = 0, /* Forever */
    DEFAULT_TX_RATE        = 100,
    DEFAULT_PRIME_COUNT    = 1,
    DEFAULT_SRC_PORT       = 1234,
    DEFAULT_DST_PORT       = 5678,
    DEFAULT_TTL            = 64,
    DEFAULT_TCP_SEQ_NUMBER = 0x12378,
    MAX_TCP_SEQ_NUMBER     = UINT32_MAX / 8,
    DEFAULT_TCP_ACK_NUMBER = 0x12390,
    MAX_TCP_ACK_NUMBER     = UINT32_MAX / 8,
    DEFAULT_TCP_FLAGS      = ACK_FLAG,
    DEFAULT_WND_SIZE       = 8192,
    MIN_VLAN_ID            = 1,
    MAX_VLAN_ID            = 4095,
    DEFAULT_VLAN_ID        = MIN_VLAN_ID,
    MIN_COS                = 0,
    MAX_COS                = 7,
    DEFAULT_COS            = MIN_COS,
    MIN_TOS                = 0,
    MAX_TOS                = 255,
    DEFAULT_TOS            = MIN_TOS,
    MAX_ETHER_TYPE_SIZE    = 0x600,
    OVERHEAD_FUDGE_VALUE   = 50
};

typedef struct port_info_s {
    /* Packet type and information */
    struct rte_ether_addr eth_dst_addr;   /**< Destination Ethernet address */
    struct rte_ether_addr eth_src_addr;   /**< Source Ethernet address */
    rte_atomic32_t port_flags;            /**< Special send flags for ARP and other */
    rte_atomic64_t transmit_count;        /**< Packets to transmit loaded into current_tx_count */
    rte_atomic64_t current_tx_count;      /**< Current number of packets to send */
    volatile uint64_t tx_cycles;          /**< Number cycles between TX bursts */
    uint16_t pid;                         /**< Port ID value */
    uint16_t tx_burst;                    /**< Number of TX burst packets */
    uint16_t lsc_enabled;                 /**< Enable link state change */
    uint16_t rx_burst;                    /**< RX burst size */
    uint64_t tx_pps;                      /**< Transmit packets per seconds */
    uint64_t tx_count;                    /**< Total count of tx attempts */
    uint64_t delta;                       /**< Delta value for latency testing */
    double tx_rate;                       /**< Percentage rate for tx packets with fractions */
    struct rte_eth_link link;             /**< Link Information like speed and duplex */
    struct rte_eth_dev_info dev_info;     /**< PCI info + driver name */
    char user_pattern[USER_PATTERN_SIZE]; /**< User set pattern values */
    fill_t fill_pattern_type;             /**< Type of pattern to fill with */
    struct rte_eth_stats curr_stats;      /**< current port statistics */
    struct rte_eth_stats queue_stats;     /**< current port queue statistics */
    struct rte_eth_stats rate_stats;      /**< current packet rate statistics */
    struct rte_eth_stats prev_stats;      /**< previous port statistics */
    struct rte_eth_stats base_stats;      /**< base port statistics */
    pkt_stats_t pkt_stats;                /**< Statistics for a number of stats */
    pkt_sizes_t pkt_sizes;                /**< Stats for the different packet sizes */
    uint64_t max_ipackets;                /**< Max seen input packet rate */
    uint64_t max_opackets;                /**< Max seen output packet rate */
    uint64_t max_missed;                  /**< Max missed packets seen */
    uint64_t qcnt[RTE_ETHDEV_QUEUE_STAT_CNTRS];      /**< queue count */
    uint64_t prev_qcnt[RTE_ETHDEV_QUEUE_STAT_CNTRS]; /**< Previous queue count */
    FILE *pcap_file;                                 /**< PCAP file handle */
} port_info_t;

/**
 * @brief Initializes the ports for packet generation.
 *
 * This function initializes the ports required for packet generation. It sets up
 * the necessary resources and configurations for each port.
 *
 * @return 0 on success, or a negative value on error.
 *
 * @note This function should be called before any other functions related to port
 *       initialization.
 */
int init_ports(void);

#ifdef __cplusplus
}
#endif

#endif /* GPKT_PORT_H_ */
