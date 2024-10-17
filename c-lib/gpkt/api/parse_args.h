// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2023-2024 Intel Corporation

#ifndef _PARSE_ARGS_H_
#define _PARSE_ARGS_H_

#include <stdbool.h>
#include <getopt.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct parse_args_s {
    bool promiscuous_mode;        // Enable promiscuous mode
    bool verbose_mode;            // Enable verbose mode

} parse_args_t;

/**
 * @brief Parses command-line arguments and initializes the packet generator.
 *
 * This function processes the command-line arguments provided by the user and
 * initializes the packet generator accordingly. It is responsible for setting up
 * the necessary configurations and options for packet generation.
 *
 * @param argc The number of command-line arguments.
 * @param argv An array of strings containing the command-line arguments.
 *
 * @return 0 on success, or a negative value on error.
 */
int parse_args(int argc, char **argv);

#ifdef __cplusplus
}
#endif

#endif /* _PARSE_ARGS_H_ */
