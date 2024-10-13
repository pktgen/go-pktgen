/*-
 * Copyright(c) <2012-2023>, Intel Corporation. All rights reserved.
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

#include <stdio.h>
#include <unistd.h>

#include <tlog.h>

#include "api_private.h"

static gpkt_t gpkt_info;
gpkt_t *gpkt = &gpkt_info;

// Define a structure to hold the arguments
static struct args_t arg_data, *args = &arg_data;
static bool gpkt_exit_flag;

// Set the string at the nth position in the array using strdup to avoid DPDK's string corruption
int
gpktSetArgv(char *s)
{
    if (args) {
        if (args->argc >= ARGV_MAX_NUM)
            return -1;

        strncpy(args->argv_str[args->argc], s, ARGV_MAX_SIZE - 1);
        args->argv[args->argc] = args->argv_str[args->argc];
        args->argc++;
        free(s);
    }
    return 0;
}

static void *
_thread_func(void *arg)
{
    struct args_t *args = arg;
    int err;

    if (pthread_setname_np(pthread_self(), "gpkt_thread"))
        TLOG_NULL_RET("Failed to set thread name\n");

    if ((err = rte_eal_init(args->argc, args->argv)) < 0)
        TLOG_NULL_RET("Error with EAL initialization Error: %d\n", rte_errno);

    if (pthread_barrier_wait(&args->barrier) > 0)
        TLOG_NULL_RET("Failed to wait for barrier\n");

    TLOG_PRINT("DPDK initializing is done, available ports %d of %d total, pid %d tid %d\n",
               rte_eth_dev_count_avail(), rte_eth_dev_count_total(), getpid(), gettid());

    gpkt_exit_flag = false;
    for (;;) {
        if (gpkt_exit_flag)
            break;
        usleep(500);
    }

    return NULL;
}

// Initialize DPDK
int
gpktStart(char *pts)
{
    if (strlen(pts) > 0 && tlog_open(pts) < 0) {
        fprintf(stderr, "%s: Failed to open log file\n", __func__);
        return -1;
    }

    if (getuid() != 0)
        TLOG_ERR_RET("Go-Pktgen must be run as root for DPDK\n");

    if (pthread_barrier_init(&args->barrier, NULL, 2))
        TLOG_ERR_RET("Failed to initialize barrier\n");

    if (pthread_create(&gpkt->pid, NULL, &_thread_func, (void *)&arg_data) == 0) {
        if (pthread_barrier_wait(&args->barrier) > 0)
            TLOG_ERR_GOTO(error, "Failed to wait on barrier\n");
    }

error:
    if (pthread_barrier_destroy(&args->barrier))
        TLOG_ERR_RET("Failed to destroy barrier\n");

    return 0;
}

void
gpktStop(void)
{
    if (gpkt) {
        gpkt_exit_flag = true;

        pthread_join(gpkt->pid, NULL);

        tlog_close();

        gpkt = NULL;
    }
}
