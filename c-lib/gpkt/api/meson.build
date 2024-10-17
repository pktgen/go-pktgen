# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2023-2024 Intel Corporation

sources = files(
  '_inet.c',
  'api.c',
  'parse_args.c',
  'port.c'
)
headers = files(
    'api.h',
    'port.h',
    'stats.h',
)

libapi = library(libname, sources, install:true, dependencies: [includes, dpdk, tlog, modes])
api = declare_dependency(link_with: libapi, include_directories: include_directories('.'))

gpkt_libs += api
