// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

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

#define TLOG_PATH_PREFIX   "/dev/pts/"
#define TLOG_PATH_MAX_SIZE 128
static char tlog_path[TLOG_PATH_MAX_SIZE];        // Path to the tlog file

int
tlog_set_path(const char *path)
{
    if (path && strlen(path) < TLOG_PATH_MAX_SIZE) {
        if (path[0] != '/') {
            strcpy(tlog_path, TLOG_PATH_PREFIX);
            strncat(tlog_path, path, sizeof(tlog_path) - strlen(TLOG_PATH_PREFIX) - 1);
        } else
            strncpy(tlog_path, path, sizeof(tlog_path) - 1);
        return 0;
    }
    return -1;
}

const char *
tlog_get_path(void)
{
    return tlog_path;
}

int
tlog_open(char *log_path)
{
    if (log_path)
        tlog_set_path(log_path);
    if (strlen(tlog_path) == 0) {
        fprintf(stderr, "Log file path not set\n");
        return 0;
    }

    if (tlog == NULL) {
        memset(&tlog_info, 0, sizeof(tlog_info));
        tlog     = &tlog_info;
        tlog->fd = -1;
    } else if (tlog->fd > 0)
        close(tlog->fd);

    tlog->fd = open(tlog_path, O_WRONLY);
    if (tlog->fd < 0) {
        fprintf(stderr, "Failed to open log file: (%s), %s(%d)\n", tlog_path, strerror(errno),
                errno);
        return -1;
    }

    return 0;
}

void
tlog_close(void)
{
    if (tlog && tlog->fd < 0)
        close(tlog->fd);

    memset(&tlog_info, 0, sizeof(tlog_info));
    tlog_info.fd = -1;
    tlog         = NULL;
}

int
tlog_vlog(const char *func, int line, const char *format, va_list ap)
{
    char buff[TLOG_BUF_SIZE + 1];
    int n;

    memset(buff, 0, sizeof(buff));
    if (func)
        snprintf(buff, TLOG_BUF_SIZE, "[%-32s:%4d] %s", func, line, format);
    else
        snprintf(buff, TLOG_BUF_SIZE, "%s", format);

    if (!tlog || tlog->fd <= 0) {
        fprintf(stderr, buff, ap);
        return 0;
    }

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
tlog_printf(const char *fn, int ln, const char *format, ...)
{
    va_list ap;
    int ret;

    va_start(ap, format);
    ret = tlog_vlog(fn, ln, format, ap);
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

    tlog_printf(__func__, __LINE__, "Stack Frames:\n");
    while (size > 0) {
        tlog_printf(__func__, __LINE__, "  %d: %s\n", size, symb[size - 1]);
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

    tlog_printf(__func__, __LINE__, "*** PANIC:\n");
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
