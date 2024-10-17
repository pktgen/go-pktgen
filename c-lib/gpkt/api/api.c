// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2023-2024 Intel Corporation

#include <stdio.h>
#include <unistd.h>

#include <gpkt.h>

#include <port.h>
#include <single.h>
#include <pcap.h>

#include "api_private.h"

static gpkt_t gpkt_info;
gpkt_t *gpkt = &gpkt_info;

// Define a structure to hold the arguments
static struct args_t arg_data, *args = &arg_data;
static bool gpkt_exit_flag;

int
gpkt_add_argv(char *arg)
{
    if (args) {
        if (args->argc >= ARGV_MAX_NUM)
            return -1;

        strncpy(args->argv_str[args->argc], arg, ARGV_MAX_SIZE - 1);
        args->argv[args->argc] = args->argv_str[args->argc];
        args->argc++;
    }
    return 0;
}

static void *
_thread_func(void *arg)
{
    struct args_t *args = arg;
    int err;

    tlog_printf("Initializing Go-Pktgen thread...\n");

    if (pthread_setname_np(pthread_self(), "gpkt_thread"))
        TLOG_NULL_RET("Failed to set thread name\n");

    tlog_printf("Initializing Go-Pktgen thread with %d args...\n", args->argc);
    for (int i = 0; i < args->argc; i++)
        tlog_printf("    argv[%d]: %s\n", i, args->argv[i]);

    if ((err = rte_eal_init(args->argc, args->argv)) < 0)
        TLOG_NULL_RET("Error with EAL initialization Error: %d\n", rte_errno);

    if (pthread_barrier_wait(&args->barrier) > 0)
        TLOG_NULL_RET("Failed to wait for barrier\n");

    TLOG_PRINT("DPDK initializing is done, available ports %d of %d total, pid %d tid %d\n",
               rte_eth_dev_count_avail(), rte_eth_dev_count_total(), getpid(), gettid());

    if (init_ports() < 0)
        TLOG_NULL_RET("Failed to initialize ports\n");
    if (init_single_mode() < 0)
        TLOG_NULL_RET("Failed to initialize single mode\n");
    if (init_pcap_mode() < 0)
        TLOG_NULL_RET("Failed to initialize pcap mode\n");

    gpkt_exit_flag = false;
    for (;;) {
        if (gpkt_exit_flag)
            break;
        usleep(10000);
    }

    return NULL;
}

// Initialize DPDK
int
gpkt_start(void)
{
    int err;

    if (tlog_open() < 0) {
        printf("Failed to open tlog (%s)\n", tlog_get_path());
        return -1;
    }

    if (getuid() != 0)
        TLOG_ERR_RET("Go-Pktgen must be run as root for DPDK\n");

    if (pthread_barrier_init(&args->barrier, NULL, 2))
        TLOG_ERR_RET("Failed to initialize barrier\n");

    if ((err = pthread_create(&gpkt->pid, NULL, &_thread_func, (void *)&arg_data)) != 0)
        TLOG_ERR_GOTO(error, "Failed to create thread error(%d)\n", err);
    else {
        tlog_printf("Go-Pktgen thread created successfully, pid %d tid %ld\n", getpid(), gpkt->pid);

        if ((err = pthread_barrier_wait(&args->barrier)) > 0)
            TLOG_ERR_GOTO(error, "Failed to wait on barrier error (%d)\n", err);

        if ((err = pthread_barrier_destroy(&args->barrier)) > 0)
            TLOG_ERR_RET("Failed to destroy barrier error (%d)\n", err);
    }

    return 0;

error:
    if ((err = pthread_barrier_destroy(&args->barrier)) > 0)
        TLOG_ERR_RET("Failed to destroy barrier error (%d)\n", err);

    return -1;
}

void
gpkt_stop(void)
{
    if (gpkt) {
        int err;

        gpkt_exit_flag = true;

        if ((err = pthread_join(gpkt->pid, NULL)) > 0)
            TLOG_RET("Failed to join thread error (%d)\n", err);

        tlog_close();

        gpkt = NULL;
    }
}
