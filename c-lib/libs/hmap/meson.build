# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2023-2025 Intel Corporation

sources = files(
    'hmap.c',
)
headers = files(
    'hmap.h',
)

libhmap = library(libname, sources, install:true, dependencies: [includes])
hmap = declare_dependency(link_with: libhmap, include_directories: include_directories('.'))

gpkt_libs += hmap
