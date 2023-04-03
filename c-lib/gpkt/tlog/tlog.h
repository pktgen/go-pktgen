/*-
 * Copyright(c) <2022-2024>, Intel Corporation. All rights reserved.
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

#ifndef TLOG_H_
#define TLOG_H_

#ifdef __cplusplus
extern "C" {
#endif

#include <stdio.h>

#define TLOG_BUF_SIZE 1024

typedef struct tlog_s {
    int fd;
    char buf[TLOG_BUF_SIZE];
    int buf_len;
} tlog_t;

int tlog_open(char *pts);
void tlog_close(void);

/**
 * Generates a log message.
 *
 * The message will be sent to stdout.
 *
 * The preferred alternative is the TLOG_LOG() macro because it adds the
 * function name and line number
 * automatically.
 *
 * @param func
 *   Function name.
 * @param line
 *   Line Number.
 * @param format
 *   The format string, as in printf(3), followed by the variable arguments
 *   required by the format.
 * @return
 *   - The number of characters printed on success.
 *   - A negative value on error.
 */
int tlog_log(const char *func, int line, const char *format, ...)
#ifdef __GNUC__
#if (__GNUC__ > 4 || (__GNUC__ == 4 && __GNUC_MINOR__ > 2))
    __attribute__((cold))
#endif
#endif
    __attribute__((format(printf, 3, 4)));

/**
 * Generates a log message regardless of log level.
 *
 * The message will be sent to stdout.
 *
 * @param format
 *   The format string, as in printf(3), followed by the variable arguments
 *   required by the format.
 * @return
 *   - The number of characters printed on success.
 *   - A negative value on error.
 */
int tlog_printf(const char *format, ...)
#ifdef __GNUC__
#if (__GNUC__ > 4 || (__GNUC__ == 4 && __GNUC_MINOR__ > 2))
    __attribute__((cold))
#endif
#endif
    __attribute__((format(printf, 1, 2)));

/**
 * Generates a log message.
 *
 * The message will be sent to stdout.
 *
 * The preferred alternative is the TLOG_LOG() macro because it adds the
 * function name and line number automatically.
 *
 * @param func
 *   Function name.
 * @param line
 *   Line Number.
 * @param format
 *   The format string, as in printf(3), followed by the variable arguments
 *   required by the format.
 * @param ap
 *   The va_list of the variable arguments required by the format.
 * @return
 *   - The number of characters printed on success.
 *   - A negative value on error.
 */
int tlog_vlog(const char *func, int line, const char *format, va_list ap)
    __attribute__((format(printf, 3, 0)));

/**
 * Generates a log message.
 *
 * The TLOG_LOG() macro is a helper that prefixes the string with the log level,
 * function name, line number, and calls tlog_log().
 *
 * @param ...
 *   The fmt string, as in printf(3), followed by the variable arguments
 *   required by the format.
 * @return
 *   - The number of characters printed on success.
 *   - A negative value on error.
 */
#define TLOG_LOG(f, ...) tlog_log(__func__, __LINE__, #f ": " __VA_ARGS__)
#define TLOG_ERR(...)    TLOG_LOG(ERR, __VA_ARGS__)

/**
 * Generate an Error log message and return value
 *
 * Same as TLOG_LOG(ERR,...) define, but returns -1 to enable this style of coding.
 *   if (val == error) {
 *       TLOG_ERR("Error: Failed\n");
 *       return -1;
 *   }
 * Returning _val  to the calling function.
 */
#define TLOG_ERR_RET_VAL(_val, ...) \
    do {                            \
        TLOG_ERR(__VA_ARGS__);      \
        return _val;                \
    } while ((0))

/**
 * Generate an Error log message and return
 *
 * Same as TLOG_LOG(ERR,...) define, but returns to enable this style of coding.
 *   if (val == error) {
 *       TLOG_ERR("Error: Failed\n");
 *       return;
 *   }
 * Returning to the calling function.
 */
#define TLOG_RET(...) TLOG_ERR_RET_VAL(, __VA_ARGS__)

/**
 * Generate an Error log message and return -1
 *
 * Same as TLOG_LOG(ERR,...) define, but returns -1 to enable this style of coding.
 *   if (val == error) {
 *       TLOG_ERR("Error: Failed\n");
 *       return -1;
 *   }
 * Returning a -1 to the calling function.
 */
#define TLOG_ERR_RET(...) TLOG_ERR_RET_VAL(-1, __VA_ARGS__)

/**
 * Generate an Error log message and return NULL
 *
 * Same as TLOG_LOG(ERR,...) define, but returns NULL to enable this style of coding.
 *   if (val == error) {
 *       TLOG_ERR("Error: Failed\n");
 *       return NULL;
 *   }
 * Returning a NULL to the calling function.
 */
#define TLOG_NULL_RET(...) TLOG_ERR_RET_VAL(NULL, __VA_ARGS__)

/**
 * Generate a Error log message and goto label
 *
 * Same as TLOG_LOG(ERR,...) define, but goes to a label to enable this style of coding.
 *   if (error condition) {
 *       TLOG_ERR("Error: Failed\n");
 *       goto lbl;
 *   }
 */
#define TLOG_ERR_GOTO(lbl, ...) \
    do {                        \
        TLOG_ERR(__VA_ARGS__);  \
        goto lbl;               \
    } while ((0))
/**
 * Generates a log message.
 *
 * The TLOG_LOG() macro is a helper that prefixes the string with the,
 * function name, line number, and calls tlog_log().
 *
 * @param ...
 *   The fmt string, as in printf(3), followed by the variable arguments
 *   required by the format.
 * @return
 *   - The number of characters printed on success.
 *   - A negative value on error.
 */
#define TLOG_LOG(f, ...) tlog_log(__func__, __LINE__, #f ": " __VA_ARGS__)

/**
 * Generates a log message regardless of log level.
 *
 * @param f
 *   The fmt string, as in printf(3), followed by the variable arguments
 *   required by the format.
 * @param args
 *   Variable arguments depend on Application.
 * @return
 *   - The number of characters printed on success.
 *   - A negative value on error.
 */
#define TLOG_PRINT(f, args...) tlog_printf(f, ##args)

#ifdef __cplusplus
}
#endif

#endif /* TLOG_H_ */
