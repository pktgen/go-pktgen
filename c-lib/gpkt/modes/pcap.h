// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

#ifndef GPKT_PCAP_H_
#define GPKT_PCAP_H_

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @brief Initializes the pcap mode for packet generation.
 *
 * This function sets up the necessary environment for packet generation using
 * the pcap library. It opens the pcap device, configures the capture and injection
 * parameters, and prepares the packet generation pipeline.
 *
 * @return 0 on success, or a negative value on error.
 */
int init_pcap_mode(void);

#ifdef __cplusplus
}
#endif

#endif /* GPKT_PCAP_H_ */
