// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2023-2024 Intel Corporation

#ifndef GPKT_API_H_
#define GPKT_API_H_

#ifdef __cplusplus
extern "C" {
#endif
/**
 * @brief Starts the gpktApi library.
 *
 * This function initializes the gpktApi library and performs any necessary setup.
 * It should be called before any other gpktApi functions are used.
 *
 * @return 0 on success, or a negative value on error.
 */
int gpkt_start(void);

/**
 * @brief Stops the gpktApi library.
 *
 * @returns N/A
 */
void gpkt_stop(void);

/**
 * @brief Adds an argument for the gpktApi library.
 *
 * @param arg The argument to be added.
 * @return 0 on success, or a -1 on error when too many arguments are provided.
 */
int gpkt_add_argv(char *arg);

#ifdef __cplusplus
}
#endif

#endif /* GPKT_API_H_ */
