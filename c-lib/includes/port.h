// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

#ifndef GPKT_PORT_H_
#define GPKT_PORT_H_

#include <stdint.h>

#include <rte_config.h>
#include <rte_ether.h>
#include <rte_ethdev.h>
#include <rte_dev.h>

#if defined(RTE_LIBRTE_PMD_BOND) || defined(RTE_NET_BOND)
#include <rte_eth_bond_8023ad.h>
#endif
#include <rte_bus_pci.h>
#include <rte_bus.h>

#include <gpkt.h>
#include <_inet.h>
#include <stats.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    uint32_t ip_addr;
    uint32_t netmask;
} ip_config_t;

#define USER_PATTERN_SIZE 16
typedef enum {
    ZERO_FILL_PATTERN = 1,
    ABC_FILL_PATTERN,
    USER_FILL_PATTERN,
    NO_FILL_PATTERN,
} fill_t;

enum {
    DEFAULT_MBUFS_PER_PORT = (32 * 1024),
    DEFAULT_CACHE_SIZE     = 256,
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
    OVERHEAD_FUDGE_VALUE   = 50,
};

enum {
    JUMBO_PKTS_FLAG = 0x0001,        // Jumbo packet flag
};

typedef struct pkt_s {
    struct rte_ether_addr eth_dst_addr;        // Destination Ethernet address
    struct rte_ether_addr eth_src_addr;        // Source Ethernet address
    pkt_hdr_t *hdr;                            // Packet header data
    uint64_t ol_flags;                         // offload flags
    uint32_t tcp_seq;                          // TCP sequence number
    uint32_t tcp_ack;                          // TCP acknowledge number
    uint16_t sport;                            // Source port value
    uint16_t dport;                            // Destination port value
    uint16_t ethType;                          // IPv4 or IPv6
    uint16_t ipProto;                          // TCP or UDP or ICMP
    uint16_t ether_hdr_size;                   // Size of Ethernet header in packet for VLAN ID
    uint16_t pkt_size;                         // Size of packet in bytes not counting FCS
    uint8_t tcp_flags;                         // TCP flags value
    uint8_t ttl;                               // TTL value for IPv4 headers
} pkt_t __rte_cache_aligned;

typedef struct port_info_s {
    rte_atomic32_t port_flags;                 // Special send flags for ARP and other
    rte_atomic64_t transmit_count;             // Packets to transmit loaded into current_tx_count
    rte_atomic64_t current_tx_count;           // Current number of packets to send
    volatile uint64_t tx_cycles;               // Number cycles between TX bursts
    uint16_t flags;                            // Special send flags
    uint16_t pid;                              // Port ID value
    uint16_t sid;                              // Socket ID value
    uint16_t rxqcnt;                           // Rx Queue count
    uint16_t txqcnt;                           // Tx Queue count
    uint16_t nb_rxd;                           // Number of RX descriptors
    uint16_t nb_txd;                           // Number of TX descriptors
    uint16_t tx_burst;                         // Number of TX burst packets
    uint16_t lsc_enabled;                      // Enable link state change
    uint16_t rx_burst;                         // RX burst size
    uint32_t cache_size;                       // Cache size for RX and TX buffers
    uint32_t nb_mbufs_per_port;                // Number of mbufs per port
    uint64_t tx_pps;                           // Transmit packets per seconds
    uint64_t tx_count;                         // Total count of tx attempts
    uint64_t delta;                            // Delta value for latency testing
    double tx_rate;                            // Percentage rate for tx packets with fractions
    struct rte_ether_addr eth_dst_addr;        // Destination Ethernet address
    struct rte_ether_addr eth_src_addr;        // Source Ethernet address
    struct rte_eth_link link;                  // Link Information like speed and duplex
    struct rte_eth_dev_info dev_info;          // PCI info + driver name
    pkt_stats_t pkt_stats;                     // Statistics for a number of stats
    // uint64_t max_ipackets;                     // Max seen input packet rate
    // uint64_t max_opackets;                     // Max seen output packet rate
    // uint64_t max_missed;                       // Max missed packets seen
    // uint64_t qcnt[RTE_ETHDEV_QUEUE_STAT_CNTRS];             // queue count
    // uint64_t prev_qcnt[RTE_ETHDEV_QUEUE_STAT_CNTRS];        // Previous queue count
    char user_pattern[USER_PATTERN_SIZE];        // User set pattern values
    fill_t fill_pattern_type;                    // Type of pattern to fill with
    FILE *pcap_file;                             // PCAP file handle
    struct rte_mempool *rx_mp;                   // Memory pool for RX packets
    pkt_t *pkt;                                  // Packet data
} port_info_t __rte_cache_aligned;

typedef struct port_config_s {         // Must match Go gpcommon.PortConfig structure
    uint16_t pid;                      // Port ID value
    uint16_t rxqcnt;                   // Rx Queue count
    uint16_t txqcnt;                   // Tx Queue count
    uint16_t nb_rxd;                   // Number of RX descriptors
    uint16_t nb_txd;                   // Number of TX descriptors
    uint16_t rx_burst;                 // RX burst size
    uint16_t tx_burst;                 // Number of TX burst packets
    uint16_t cache_size;               // Cache size for RX and TX buffers
    uint32_t nb_mbufs_per_port;        // Number of mbufs per port
} port_config_t;

#define INFO_NAME_SIZE 32
typedef struct device_info_s {
    char name[INFO_NAME_SIZE];             // Device name
    char bus_name[INFO_NAME_SIZE];         // Bus name
    struct rte_ether_addr mac_addr;        // MAC address
    uint32_t if_index;                     // Interface index
    uint32_t min_mtu;                      // Minimum MTU value
    uint32_t max_mtu;                      // Maximum MTU value
    uint32_t min_rx_bufsize;               // Minimum RX buffer size
    uint32_t max_rx_bufsize;               // Maximum RX buffer size
    uint32_t max_rx_pktlen;                // Maximum RX packet length
    uint32_t max_rx_queues;                // Maximum number of RX queues
    uint32_t max_tx_queues;                // Maximum number of TX queues
    uint32_t max_mac_addrs;                // Maximum number of MAC addresses
    uint32_t max_hash_mac_addrs;           // Maximum number of hash addresses
    uint32_t max_vfs;                      // Maximum number of VFs
    uint32_t nb_rx_queues;                 // Number of RX queues
    uint32_t nb_tx_queues;                 // Number of TX queues
    uint32_t socket_id;                    // NUMA node index
} device_info_t;

/**
 * @brief Initializes a port for packet generation.
 *
 * This function initializes the ports required for packet generation. It sets up
 * the necessary resources and configurations for each port.
 *
 * @param g Global packet generator context.
 * @param pid Port ID value.
 * @param rx_qid Queue ID value.
 * @param tx_qid Queue ID value.
 *
 * @return 0 on success, or a negative value on error.
 *
 * @note This function should be called before any other functions related to port
 *       initialization.
 */
GPKT_API int port_init(uint16_t pid);

GPKT_API void *port_rxtx_loop(gpkt_t *g, uint16_t pid, uint16_t rx_qid, uint16_t tx_qid);

GPKT_API int port_set_info(port_config_t *cfg);

GPKT_API port_config_t *port_get_info(uint16_t port_id);

GPKT_API port_info_t *port_info_get(uint16_t port_id);

GPKT_API void port_info_free(port_config_t *cfg);

static __inline__ uint32_t
lport_encode(uint16_t pid, uint16_t qid)
{
    return (pid << 16) | qid;
}

static __inline__ void
lport_decode(uint32_t lport, uint16_t *pid, uint16_t *qid)
{
    if (pid)
        *pid = (lport >> 16) & 0xFFFF;
    if (qid)
        *qid = lport & 0xFFFF;
}

#ifdef __cplusplus
}
#endif

#endif /* GPKT_PORT_H_ */
