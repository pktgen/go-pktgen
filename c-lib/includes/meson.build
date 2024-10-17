# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2023-2024 Intel Corporation

sources = []
headers = files(
    '_inet.h',
    'constants.h',
)

includes = declare_dependency(include_directories: include_directories('.'))

install_headers(headers, subdir: meson.project_name().to_lower())
