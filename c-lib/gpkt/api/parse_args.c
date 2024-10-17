// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2023-2024 Intel Corporation

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include <tlog.h>
#include <parse_args.h>

static const char *short_options = "m:Pvh";

#define MAPPING_OPT     "map"
#define PROMISCUOUS_OPT "promiscuous"
#define VERBOSE_OPT     "verbose"
#define HELP_OPT        "help"

// clang-format off
static const struct option lgopts[] = {
    {MAPPING_OPT,           1, 0, 'm'},
    {PROMISCUOUS_OPT,       0, 0, 'P'},
    {VERBOSE_OPT,           0, 0, 'v'},
    {HELP_OPT,              0, 0, 'h'},
    {NULL,                  0, 0, 0}
};
// clang-format on

static parse_args_t info;

/* display usage */
static void
usage(int err)
{
    printf("pktgen [EAL options] -- [-m map] [-P] [-h]\n"
           "\t-m|--map <map>           Core to Port/queue mapping '[Rx-Cores:Tx-Cores].port'\n"
           "\t-P|--no-promiscuous      Turn off promiscuous mode (default On)\n"
           "\t-h|--help                Print this help\n");

    exit(err);
}

int
parse_add_map(char *map)
{
    (void)map;
    return 0;        // Implement parsing logic for map option here
}

// Parse command-line arguments for pktgen
int
parse_args(int argc, char **argv)
{
    int opt;
    char **argvopt;
    int option_index;

    argvopt = argv;

    info.promiscuous_mode = true;

    tlog_printf("%s: started\n", __func__);

    if (argc <= 0)
        return 0;

    // Implement parsing logic for command-line arguments here
    while ((opt = getopt_long(argc, argvopt, short_options, lgopts, &option_index)) != EOF) {
        switch (opt) {
        case 'm':
            parse_add_map(optarg);
            break;
        case 'P':
            info.promiscuous_mode = false;
            break;
        case 'v':
            info.verbose_mode = true;
            break;
        case 'h':
            usage(0);
            break;
        default:
            usage(1);
        }
    }

    return 0;
}
