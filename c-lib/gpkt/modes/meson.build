# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2023-2025 Intel Corporation

sources = files(
  'single.c',
  'pcap.c',
)
headers = files(
    'pcap.h',
    'single.h',
)

libmodes = library(libname, sources, install:true, dependencies: [includes, dpdk, tlog])
modes = declare_dependency(link_with: libmodes, include_directories: include_directories('.'))

gpkt_libs += modes
