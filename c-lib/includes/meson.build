# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2023-2025 Intel Corporation

sources = []
headers = files(
    '_inet.h',
    'dpdk_api.h',
    'gapi_mutex_helper.h',
    'gpkt.h',
    'messages.h',
    'msgchan.h',
    'port.h',
    'stats.h',
)

includes = declare_dependency(include_directories: include_directories('.'))

install_headers(headers, subdir: meson.project_name().to_lower())
