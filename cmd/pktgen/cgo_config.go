/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package main

/*
#cgo pkg-config: libdpdk
#cgo CFLAGS: -I../../c-lib/usr/local/include/go-pktgen
#cgo LDFLAGS: -L../../c-lib/usr/local/lib/x86_64-linux-gnu -lgpkt -lbsd
*/
import "C"
