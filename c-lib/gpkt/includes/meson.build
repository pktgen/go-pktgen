# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2022-2024 Intel Corporation

sources = []
headers = files(
  'gpkt_common.h',
)

includes = declare_dependency(include_directories: include_directories('.'))

install_headers(headers, subdir: meson.project_name().to_lower())
