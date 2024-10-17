// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2023-2024 Intel Corporation

#ifndef API_PRIVATE_H_
#define API_PRIVATE_H_

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <fcntl.h>

#include <pthread.h>

#include <rte_eal.h>
#include <rte_errno.h>
#include <rte_ethdev.h>

#ifdef __cplusplus
extern "C" {
#endif

#define ARGV_MAX_NUM  64
#define ARGV_MAX_SIZE 128

// Local structure to hold command-line arguments and other application-specific data
struct args_t {
    int argc;
    char *argv[ARGV_MAX_NUM];
    char argv_str[ARGV_MAX_NUM][ARGV_MAX_SIZE];        // Concatenated command-line arguments
    pthread_barrier_t barrier;
};

typedef struct gpkt_s {
    pthread_t pid;                  // Process ID
    volatile bool exit_flag;        // Flag to indicate whether to exit the application
} gpkt_t;

extern gpkt_t *gpkt;

#ifdef __cplusplus
}
#endif

#endif /* API_PRIVATE_H_ */
