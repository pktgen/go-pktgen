# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2022-2024 Intel Corporation

subdir('includes')

dirs = [
    'single',
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

