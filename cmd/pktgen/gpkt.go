// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <rte_eal.h>

static char **
makeCharArray(int size)
{
    return calloc(sizeof(char*), size);
}

static void
setArrayString(char **a, char *s, int n)
{
    a[n] = strdup(s); // Need to allocate memory for the string as DPDK corrupts otherwise
}

static void
freeCharArray(char **a, int size)
{
	for (int i = 0; i < size; i++) {
		if (a[i] && strlen(a[i]) > 0)
			free(a[i]);
	}
	free(a);
}

static int
gpkt_init(int argc, char **argv)
{
	int err;

    if ((err = rte_eal_init(argc, argv)) < 0) {
        printf("Error with EAL initialization %d\n", err);
        return -1;
    }

    return 0;
}
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/pktgen/go-pktgen/internal/cfg"
)

// gpktInit returning the basic information string
func gpktInit(cs *cfg.System) error {

	argv, err := cs.MakeArgs()
	if err != nil {
		return err
	}

	argc := len(argv)
	cargs := C.makeCharArray(C.int(argc) + 1)
	args := make([]*C.char, argc)

	for i, s := range argv {
		cStr := C.CString(s)
		C.setArrayString(cargs, cStr, C.int(i))
		args = append(args, cStr)
	}
	defer func() {
		// Free the strdup() strings from the setArrayString() call
		C.freeCharArray(cargs, C.int(argc))

		// Now we free the C strings created by C.CString() calls
		for _, a := range args {
			if a != nil && C.GoString(a) != "" {
				C.free(unsafe.Pointer(a))
			}
		}
	}()

	pktgen.dbg.Printf("argc %d, argv %v\n", argc, argv)

	// It appears DPDK rte_eal_init() corrupted the argc/argv values,
	// so we needed to jump through a few hoops to get it to work.
	// We get a double free error if we don't jump through the hoops.
	if ret := C.gpkt_init(C.int(argc), cargs); ret < 0 {
		return fmt.Errorf("gpkt_init failed")
	}
	return nil
}
