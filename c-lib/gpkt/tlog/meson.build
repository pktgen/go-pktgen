# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2022-2024 Intel Corporation

sources = files(
    'tlog.c',
)
headers = files(
    'tlog.h',
)

libtlog = library(libname, sources, install:true, dependencies: [includes, dpdk])
tlog = declare_dependency(link_with: libtlog, include_directories: include_directories('.'))

gpkt_libs += tlog
