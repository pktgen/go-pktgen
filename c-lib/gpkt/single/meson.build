# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2022-2024 Intel Corporation

sources = files(
  'single-mode.c',
)
headers = files(
    'single-mode.h',
)

libsingle = library(libname, sources, install:true, dependencies: [includes, dpdk])
single = declare_dependency(link_with: libsingle, include_directories: include_directories('.'))

gpkt_libs += single
