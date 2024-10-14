/*-
 * Copyright(c) <2022-2024>, Intel Corporation. All rights reserved.
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

#ifndef GPKT_API_H_
#define GPKT_API_H_

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @brief Initializes the gpktApi library.
 *
 * This function initializes the gpktApi library and performs any necessary setup.
 * It should be called before any other gpktApi functions are used.
 *
 * @return 0 on success, or a negative value on error.
 */
int gpktStart(void);

void gpktStop(void);

int gpktSetArgv(char *s);

#ifdef __cplusplus
}
#endif

#endif /* GPKT_API_H_ */
