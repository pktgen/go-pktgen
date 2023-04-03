# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2023-2025 Intel Corporation

sources = files(
    'tlog.c',
)
headers = files(
    'tlog.h',
)

libtlog = library(libname, sources, install:true, dependencies: [])
tlog = declare_dependency(link_with: libtlog, include_directories: include_directories('.'))

gpkt_libs += tlog
