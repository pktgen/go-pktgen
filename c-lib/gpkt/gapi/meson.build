# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2023-2025 Intel Corporation

# 'channel.c',
sources = files(
    '_inet.c',
    'dpdk_api.c',
    'messages.c',
    'msgchan.c',
    'port.c',
    'rxtx.c'
)
headers = files()

libgapi = library(libname, sources, install:true, dependencies: [includes, dpdk, tlog, modes])
gapi = declare_dependency(link_with: libgapi, include_directories: include_directories('.'))

gpkt_libs += gapi
