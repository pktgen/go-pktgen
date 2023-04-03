// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

#ifndef GPKT_SINGLE_H_
#define GPKT_SINGLE_H_

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @brief Initializes the single mode for packet generation.
 *
 * This function initializes the single mode for packet generation. In single mode,
 * only one packet is generated and sent. After the packet is sent.
 *
 * @return 0 on success, or a negative value on error.
 *
 * @note This function should be called before starting the packet generation.
 */
int init_single_mode(void);

#ifdef __cplusplus
}
#endif

#endif /* GPKT_SINGLE_H_ */
