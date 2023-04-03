// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

#include <stdio.h>
#include <unistd.h>

#include <rte_common.h>
#include <rte_launch.h>

#include <gpkt.h>

#include <port.h>
#include <single.h>
#include <pcap.h>
#include <msgchan.h>
#include <messages.h>

// Define a structure to hold the arguments
gpkt_t gpkt_data = {0}, *gpkt = &gpkt_data;
static gpkt_args_t args_data = {0}, *args = &args_data;

static void
print_command_line(int argc, char **argv)
{
    char buffer[1024] = {0};
    int n;

    n = snprintf(buffer, sizeof(buffer),
                 "Initializing Go-Pktgen thread with %d args\n    argv:", argc);
    for (int i = 0; i < argc; i++)
        n += snprintf(&buffer[n], sizeof(buffer) - n, "%s ", argv[i]);

    printf("%s\n", buffer);
}

static void *
dpdk_func(void *arg __rte_unused)
{
    int argc    = args->argc;
    char **argv = args->argv;
    int ret;

    if ((ret = pthread_detach(pthread_self())) > 0) {
        printf("%s-%d: Failed to detach EAL thread error %d\n", __func__, __LINE__, ret);
        goto leave;
    }

    if (pthread_setname_np(pthread_self(), "eal_init_thread")) {
        printf("%s-%d: Failed to set thread name\n", __func__, __LINE__);
        goto leave;
    }

    // Display the provided arguments
    print_command_line(argc, argv);

    // Initialize DPDK
    if (rte_eal_init(argc, argv) < 0) {
        printf("%s-%d: Error with EAL initialization Error: %d\n", __func__, __LINE__, rte_errno);
        goto leave;
    }

    printf("DPDK initialization is done, available ports %d of %d total, pid %d tid %d\n",
           rte_eth_dev_count_avail(), rte_eth_dev_count_total(), getpid(), gettid());

    // Initialize the control plane message channel
    if ((gpkt->dpdk_chnl = mc_create("DPDK", DEFAULT_MSGCHAN_SIZE)) == NULL)
        TLOG_ERR_GOTO(leave, "Failed to initialize DPDK channel\n");

    // Wait to signal to the main thread is ready
    if ((ret = pthread_barrier_wait(&args->barrier)) > 0)
        TLOG_ERR_GOTO(leave, "Failed to wait on barrier error (%d)\n", ret);

    printf("Start looping waiting for messages...\n");

    // Wait for messages from the control plane
    while (gpkt->quit[rte_lcore_id()] == 0) {
        if (msg_channel_process(gpkt->dpdk_chnl) < 0)
            TLOG_ERR_GOTO(leave, "Failed to process message\n");
        usleep(10000);        // Sleep for 10ms
    }
    printf("Exiting main DPDK thread...\n");

leave:
    mc_destroy(gpkt);
    rte_eal_cleanup();

    return NULL;
}

// Initialize DPDK
int
dpdk_startup(char *log_path)
{
    int ret;

    if (tlog_open(log_path) < 0)
        TLOG_ERR_RET("Failed to open tlog (%s)\n", tlog_get_path());

    TLOG_PRINT("Starting main DPDK thread...\n");

    if (getuid() != 0)
        TLOG_ERR_RET("Go-Pktgen must be run as root for DPDK\n");

    if (pthread_barrier_init(&args->barrier, NULL, 2))
        TLOG_ERR_RET("Failed to initialize barrier\n");

    // Start the main DPDK thread
    if ((ret = pthread_create(&gpkt->thread, NULL, &dpdk_func, NULL)) != 0)
        TLOG_ERR_GOTO(leave, "Failed to create thread error(%d)\n", ret);
    else {
        if ((ret = pthread_barrier_wait(&args->barrier)) > 0)
            TLOG_ERR_GOTO(leave, "Failed to wait on barrier error (%d)\n", ret);

        TLOG_PRINT("Main DPDK thread created successfully, pid %d tid %ld\n", getpid(),
                   gpkt->thread);

        if ((ret = pthread_barrier_destroy(&args->barrier)) > 0)
            TLOG_ERR_RET("Failed to destroy barrier error (%d)\n", ret);
    }

    return 0;

leave:
    if ((ret = pthread_barrier_destroy(&args->barrier)) > 0)
        TLOG_ERR_RET("Failed to destroy barrier error (%d)\n", ret);

    return -1;
}

void
dpdk_shutdown(void)
{
    TLOG_PRINT("Stop all lcores ...\n");

    for (int i = 0; i < RTE_MAX_LCORE; i++)
        gpkt->quit[i] = 1;

    tlog_close();
}

int
dpdk_add_argv(char *argv)
{
    if (args->argc >= ARGV_MAX_NUM)
        return -1;

    strncpy(args->argv_str[args->argc], argv, ARGV_MAX_SIZE - 1);
    args->argv[args->argc] = args->argv_str[args->argc];
    args->argc++;

    return 0;
}

void
dpdk_l2p_config_dump(l2p_config_t *cfg)
{
    char *mode_str[] = MODE_STRINGS;
    char buffer[2048];
    int n = 0;

    n += snprintf(&buffer[n], sizeof(buffer) - n, "L2P Config %u\n", cfg->core_id);
    n += snprintf(&buffer[n], sizeof(buffer) - n, "   Logical Port : %08x\n", cfg->lport_id);
    n += snprintf(&buffer[n], sizeof(buffer) - n, "   Mode         : %s\n", mode_str[cfg->mode]);
    n += snprintf(&buffer[n], sizeof(buffer) - n, "   RxQid        : %u\n", cfg->rx_qid);
    n += snprintf(&buffer[n], sizeof(buffer) - n, "   TxQid        : %u\n", cfg->tx_qid);
    n += snprintf(&buffer[n], sizeof(buffer) - n, "   Port ID      : %u\n", cfg->port_id);
    n += snprintf(&buffer[n], sizeof(buffer) - n, "   Num Rx Queues: %u\n", cfg->num_rx_queues);
    n += snprintf(&buffer[n], sizeof(buffer) - n, "   Num Tx Queues: %u\n", cfg->num_tx_queues);

    printf("%s", buffer);
}

int
dpdk_l2p_config(l2p_config_t *cfg)
{
    logical_core_t *lcore = NULL;
    logical_port_t *lport = NULL;
    physical_port_t *port = NULL;

    if (cfg == NULL)
        TLOG_ERR_RET("Invalid logical port configuration\n");

    printf("Configuring L2P for core %x, port %u\n", cfg->core_id, cfg->port_id);

    dpdk_l2p_config_dump(cfg);

    if (cfg->core_id >= RTE_MAX_LCORE)
        TLOG_ERR_RET("Invalid core ID %d\n", cfg->core_id);

    if (cfg->port_id >= RTE_MAX_ETHPORTS)
        TLOG_ERR_RET("Invalid port ID (%d)\n", cfg->port_id);

    // Setup physical port structure
    port = &gpkt->ports[cfg->port_id];
    if (port->port_id == RTE_MAX_ETHPORTS) {
        port->port_id       = cfg->port_id;
        port->num_rx_queues = cfg->num_rx_queues;
        port->num_tx_queues = cfg->num_tx_queues;
    }

    // Setup logical port structure
    lport = calloc(1, sizeof(logical_port_t));
    if (lport == NULL)
        TLOG_ERR_GOTO(leave, "Failed to allocate logical port\n");

    lport->physical_port = port;
    lport->lport_id      = cfg->lport_id;
    lport->rx_qid        = cfg->rx_qid;
    lport->tx_qid        = cfg->tx_qid;

    // Setup logical core structure
    lcore               = &gpkt->lcores[cfg->core_id];
    lcore->logical_port = lport;
    lcore->core_id      = cfg->core_id;
    lcore->mode         = cfg->mode;

    return 0;
leave:
    // Reset structures on failure
    port->port_id  = RTE_MAX_ETHPORTS;
    lcore->core_id = RTE_MAX_LCORE;

    return -1;
}

int
launch_func(void *arg __rte_unused)
{
    gpkt_t *g             = arg;
    logical_core_t *lcore = &gpkt->lcores[rte_lcore_id()];

    TLOG_PRINT("Go-Pktgen on lcore %u pid %d tid %d\n", rte_lcore_id(), getpid(), gettid());

    if (lcore->mode < RXONLY_MODE || lcore->mode > RXTX_MODE)
        TLOG_ERR_RET("Invalid mode %u\n", lcore->mode);

    port_rxtx_loop(g, lcore->logical_port->physical_port->port_id, lcore->logical_port->rx_qid,
                   lcore->logical_port->tx_qid);

    return 0;
}

void
dpdk_l2p_dump(uint16_t core_id)
{
    logical_core_t *lcore = &gpkt->lcores[core_id];
    char *mode_str[]      = MODE_STRINGS;
    char buffer[2048];
    int n = 0;

    if (lcore == NULL)
        TLOG_RET("Logical Core %u not found\n", core_id);

    n += snprintf(&buffer[n], sizeof(buffer) - n, "Logical Core %u\n", core_id);
    n += snprintf(&buffer[n], sizeof(buffer) - n, "   Logical Port : %p\n", lcore->logical_port);
    n += snprintf(&buffer[n], sizeof(buffer) - n, "   Mode         : %s\n", mode_str[lcore->mode]);
    n += snprintf(&buffer[n], sizeof(buffer) - n, "   Core ID      : %u\n", lcore->core_id);

    if (lcore->logical_port != NULL) {
        n += snprintf(&buffer[n], sizeof(buffer) - n, "   Physical Port: %p\n",
                      lcore->logical_port->physical_port);
        n += snprintf(&buffer[n], sizeof(buffer) - n, "   Rx Queue ID  : %u\n",
                      lcore->logical_port->rx_qid);
        n += snprintf(&buffer[n], sizeof(buffer) - n, "   Tx Queue ID  : %u\n",
                      lcore->logical_port->tx_qid);
        if (lcore->logical_port->physical_port != NULL) {
            n += snprintf(&buffer[n], sizeof(buffer) - n, "   Port ID      : %u\n",
                          lcore->logical_port->physical_port->port_id);
            n += snprintf(&buffer[n], sizeof(buffer) - n, "   Num Rx Queues: %u\n",
                          lcore->logical_port->physical_port->num_rx_queues);
            n += snprintf(&buffer[n], sizeof(buffer) - n, "   Num Tx Queues: %u\n",
                          lcore->logical_port->physical_port->num_tx_queues);
        }
    }
    TLOG_PRINT("%s", buffer);
}

RTE_INIT_PRIO(dpdk_constructor, LAST)
{
    // Initialize the ports and lcores structures
    for (uint16_t pid = 0; pid < RTE_MAX_ETHPORTS; pid++)
        gpkt->ports[pid].port_id = RTE_MAX_ETHPORTS;
    for (uint16_t lid = 0; lid < RTE_MAX_LCORE; lid++)
        gpkt->lcores[lid].core_id = RTE_MAX_LCORE;
}
