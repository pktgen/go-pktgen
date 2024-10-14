/*-
 * Copyright(c) <2012-2023>, Intel Corporation. All rights reserved.
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <execinfo.h>
#include <errno.h>

#include "tlog.h"

static tlog_t tlog_info, *tlog;

#define TLOG_PATH_PREFIX "/dev/pts/"

int
tlog_open(int pts)
{
    char buffer[64];

    if (pts == 0)
        return 0;

    if (tlog == NULL) {
        memset(&tlog_info, 0, sizeof(tlog_info));
        tlog     = &tlog_info;
        tlog->fd = -1;
    }

    if (tlog->fd > 0)
        close(tlog->fd);

    snprintf(buffer, sizeof(buffer) - 1, TLOG_PATH_PREFIX "%d", pts);

    tlog->fd = open(buffer, O_WRONLY);
    if (tlog->fd < 0) {
        fprintf(stderr, "Failed to open log file: (%s), %s(%d)\n", buffer, strerror(errno), errno);
        return -1;
    }

    printf("%s: Logging Started: '%s' @ %d\n", __func__, buffer, tlog->fd);
    tlog_printf("\n%s: Logging Started: '%s'\n", __func__, buffer);
    return 0;
}

void
tlog_close(void)
{
    if (tlog && tlog->fd < 0)
        close(tlog->fd);

    tlog = NULL;
    memset(&tlog_info, 0, sizeof(tlog_info));
    tlog_info.fd = -1;
}

int
tlog_vlog(const char *func, int line, const char *format, va_list ap)
{
    char buff[TLOG_BUF_SIZE + 1];
    int n;

    if (tlog->fd <= 0)
        return 0;

    memset(buff, 0, sizeof(buff));
    if (func)
        snprintf(buff, TLOG_BUF_SIZE, "(%-24s:%4d) %s", func, line, format);
    else
        snprintf(buff, TLOG_BUF_SIZE, "%s", format);

        /* GCC allows the non-literal "buff" argument whereas clang does not */
#ifdef __clang__
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wformat-nonliteral"
#endif /* __clang__ */
    n = vdprintf(tlog->fd, buff, ap);
    if (n < 0) {
        fprintf(stderr, "ERROR: Failed to write to log file: %s(%d)\n", strerror(errno), errno);
        return n;
    }

    return n;
#ifdef __clang__
#pragma clang diagnostic pop
#endif /* __clang__ */
}

int
tlog_log(const char *func, int line, const char *format, ...)
{
    va_list ap;
    int ret;

    va_start(ap, format);
    ret = tlog_vlog(func, line, format, ap);
    va_end(ap);

    return ret;
}

int
tlog_printf(const char *format, ...)
{
    va_list ap;
    int ret;

    va_start(ap, format);
    ret = tlog_vlog(NULL, 0, format, ap);
    va_end(ap);

    return ret;
}

#define BACKTRACE_SIZE 256

/* dump the stack of the calling core */
void
tlog_dump_stack(void)
{
    void *func[BACKTRACE_SIZE];
    char **symb = NULL;
    int size;

    size = backtrace(func, BACKTRACE_SIZE);
    symb = backtrace_symbols(func, size);

    if (symb == NULL)
        return;

    tlog_printf("Stack Frames:\n");
    while (size > 0) {
        tlog_printf("  %d: %s\n", size, symb[size - 1]);
        size--;
    }
    fflush(stdout);

    free(symb);
}

/* call abort(), it will generate a coredump if enabled */
void
__tlog_panic(const char *funcname, int line, const char *format, ...)
{
    va_list ap;

    tlog_printf("*** PANIC:\n");
    va_start(ap, format);
    tlog_vlog(funcname, line, format, ap);
    va_end(ap);

    tlog_dump_stack();
    abort();
}

void
__tlog_exit(const char *func, int line, const char *format, ...)
{
    va_list ap;

    va_start(ap, format);
    tlog_vlog(func, line, format, ap);
    va_end(ap);

    exit(-1);
}
