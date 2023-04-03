# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2023-2025 Intel Corporation

dirs = [
    'tlog',
    'hmap',
]

foreach d:dirs
    deps = [build_cfg]

    sources = []
    headers = []

    libname = 'gpkt_' + d
    enabled_libs += libname

    subdir(d)

    install_headers(headers, subdir: meson.project_name().to_lower())
endforeach
