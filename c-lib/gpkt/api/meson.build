# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2022-2024 Intel Corporation

sources = files(
  'api.c',
)
headers = files(
    'api_private.h',
)

libapi = library(libname, sources, install:true, dependencies: [includes, dpdk, tlog])
api = declare_dependency(link_with: libapi, include_directories: include_directories('.'))

gpkt_libs += api
