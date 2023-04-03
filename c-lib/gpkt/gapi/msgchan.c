/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

#include <inttypes.h>
#include <bsd/string.h>
#include <bsd/sys/queue.h>

#include <errno.h>
#include <stdio.h>
#include <string.h>
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <sys/types.h>
#include <pthread.h>

#include <rte_common.h>
#include <rte_log.h>
#include <rte_spinlock.h>
#include <rte_cycles.h>

#include <gpkt.h>
#include <gapi_mutex_helper.h>
#include <msgchan_priv.h>
#include <msgchan.h>

TAILQ_HEAD(msgchan_list, msg_chan);

static struct msgchan_list mc_list_head = TAILQ_HEAD_INITIALIZER(mc_list_head);
static pthread_mutex_t mc_list_mutex;

static inline void
mc_list_lock(void)
{
    int ret = pthread_mutex_lock(&mc_list_mutex);

    if (ret)
        TLOG_PRINT("failed: %s\n", strerror(ret));
}

static inline void
mc_list_unlock(void)
{
    int ret = pthread_mutex_unlock(&mc_list_mutex);

    if (ret)
        TLOG_PRINT("failed: %s\n", strerror(ret));
}

static inline void
mc_child_lock(msg_chan_t *mc)
{
    int ret = pthread_mutex_lock(&mc->mutex);

    if (ret)
        TLOG_PRINT("failed: %s\n", strerror(ret));
}

static inline void
mc_child_unlock(msg_chan_t *mc)
{
    int ret = pthread_mutex_unlock(&mc->mutex);

    if (ret)
        TLOG_PRINT("failed: %s\n", strerror(ret));
}

msgchan_t *
mc_create(const char *name, uint32_t sz)
{
    msg_chan_t *mc;
    char rname[RTE_RING_NAMESIZE + 1];

    TLOG_PRINT("Creating msg_chan_t: %s\n", name);

    /* Make sure the name is not already used or needs child created */
    mc_list_lock();
    mc = mc_lookup(name);
    if (mc)
        TLOG_ERR_GOTO(err, "msgchan_t with the same name already exists\n");

    mc = calloc(1, sizeof(msg_chan_t));
    if (!mc)
        TLOG_ERR_GOTO(err, "Failed to allocate memory\n");

    snprintf(mc->name, sizeof(mc->name), "%s", name);

    snprintf(rname, sizeof(rname), "Rx:%s", name);        // Rx - Receive Ring
    if ((mc->rings[MC_RECV_RING] =
             rte_ring_create_elem(rname, RTE_CACHE_LINE_SIZE, sz, rte_socket_id(), 0)) == NULL)
        TLOG_ERR_GOTO(err, "Failed to create Recv ring\n");

    snprintf(rname, sizeof(rname), "Tx:%s", name);        // Tx - Receive Ring
    if ((mc->rings[MC_SEND_RING] =
             rte_ring_create_elem(rname, RTE_CACHE_LINE_SIZE, sz, rte_socket_id(), 0)) == NULL)
        TLOG_ERR_GOTO(err, "Failed to create Send ring\n");

    if (gapi_mutex_create(&mc->mutex, PTHREAD_MUTEX_RECURSIVE_NP))
        TLOG_ERR_GOTO(err, "creating recursive mutex failed\n");

    mc->cookie       = MC_COOKIE;
    mc->mutex_inited = true;

    TAILQ_INIT(&mc->children);

    TAILQ_INSERT_TAIL(&mc_list_head, mc, next);

    mc_list_unlock();

    return mc;
err:
    if (mc) {
        rte_ring_free(mc->rings[MC_RECV_RING]);
        rte_ring_free(mc->rings[MC_SEND_RING]);

        if (mc->mutex_inited && gapi_mutex_destroy(&mc->mutex))
            TLOG_ERR("Failed to destroy mutex\n");
        memset(mc, 0, sizeof(msg_chan_t));
        free(mc);
    }

    mc_list_unlock();

    return NULL;
}

static msgchan_t *
attach_child(msg_chan_t *parent)
{
    msg_chan_t *child = NULL;

    mc_child_lock(parent);
    child = calloc(1, sizeof(msg_chan_t));
    if (!child) {
        mc_child_unlock(parent);
        TLOG_NULL_RET("Failed to allocate new child msg_chan_t structure\n");
    }

    snprintf(child->name, RTE_RING_NAMESIZE, "%s:%d", parent->name, parent->nchildren);

    child->parent              = parent; /* Set the parent pointer in the child */
    child->cookie              = parent->cookie;
    child->rings[MC_RECV_RING] = parent->rings[MC_SEND_RING]; /* Swap Tx/Rx rings */
    child->rings[MC_SEND_RING] = parent->rings[MC_RECV_RING];
    parent->nchildren++;

    TAILQ_INSERT_TAIL(&parent->children, child, next);
    mc_child_unlock(parent);

    return child;
}

msgchan_t *
mc_attach(const char *parent_name)
{
    msg_chan_t *parent, *child = NULL;

    mc_list_lock();
    parent = mc_lookup(parent_name);

    if (parent)
        child = attach_child(parent);

    mc_list_unlock();
    return child;
}

void
mc_destroy(msgchan_t *_mc)
{
    msg_chan_t *mc = _mc;

    if (mc && mc->cookie == MC_COOKIE) {
        mc_list_lock();
        if (!mc->parent) { /* Handle parent destroy */
            msg_chan_t *m;

            TAILQ_REMOVE(&mc_list_head, mc, next);

            rte_ring_free(mc->rings[MC_RECV_RING]);
            rte_ring_free(mc->rings[MC_SEND_RING]);

            while (!TAILQ_EMPTY(&mc->children)) {
                m = TAILQ_FIRST(&mc->children);

                TAILQ_REMOVE(&mc->children, m, next);

                memset(m, 0, sizeof(msg_chan_t));
                free(m);
            }

            memset(mc, 0, sizeof(msg_chan_t));

            if (mc->mutex_inited && gapi_mutex_destroy(&mc->mutex))
                TLOG_ERR("Failed to destroy mutex\n");

            free(mc);
        } else { /* Handle child destroy */
            mc_child_lock(mc->parent);
            TAILQ_REMOVE(&mc->parent->children, mc, next);
            mc_child_unlock(mc->parent);

            memset(mc, 0, sizeof(msg_chan_t));
            free(mc);
        }
        mc_list_unlock();
    }
}

static int
__recv(msg_chan_t *mc, void **objs, int count, uint64_t msec)
{
    struct rte_ring *r;
    int nb_objs = 0;

    mc->recv_calls++;

    if (count == 0)
        return 0;

    r = mc->rings[MC_RECV_RING];

    if (msec) {
        uint64_t begin, stop;

        begin = rte_rdtsc_precise();
        stop  = begin + ((rte_get_timer_hz() / 1000) * msec);

        while (nb_objs == 0 && begin < stop) {
            nb_objs = rte_ring_dequeue_burst_elem(r, objs, RTE_CACHE_LINE_SIZE, count, NULL);
            if (nb_objs == 0) {
                begin = rte_rdtsc_precise();
                rte_pause();
            }
        }
        if (nb_objs == 0)
            mc->recv_timeouts++;
    } else
        nb_objs = rte_ring_dequeue_burst_elem(r, objs, RTE_CACHE_LINE_SIZE, count, NULL);

    mc->recv_cnt += nb_objs;
    return nb_objs;
}

static int
__send(msgchan_t *_mc, void **objs, int count)
{
    msg_chan_t *mc = _mc;
    struct rte_ring *r;
    int nb_objs;

    mc->send_calls++;

    r = mc->rings[MC_SEND_RING];

    nb_objs = rte_ring_enqueue_burst_elem(r, objs, RTE_CACHE_LINE_SIZE, count, NULL);
    if (nb_objs < 0)
        TLOG_ERR_RET("[orange]Sending to msgchan failed[]\n");

    mc->send_cnt += nb_objs;
    return nb_objs;
}

int
mc_send(msgchan_t *_mc, void **objs, int count)
{
    msg_chan_t *mc = _mc;
    int n;

    if (!mc || !objs || mc->cookie != MC_COOKIE)
        TLOG_ERR_RET("Invalid parameters\n");

    if (count < 0)
        TLOG_ERR_RET("Count of objects is negative\n");

    n = __send(mc, objs, count);
    if (n > 0)
        TLOG_PRINT("Sent %d messages\n", n);
    return n;
}

int
mc_recv(msgchan_t *_mc, void **objs, int count, uint64_t msec)
{
    msg_chan_t *mc = _mc;
    int n;

    if (!mc || !objs || mc->cookie != MC_COOKIE)
        TLOG_ERR_RET("Invalid parameters Cookie %08x\n", mc->cookie);

    if (count < 0)
        TLOG_ERR_RET("Count of objects is %d\n", count);

    n = __recv(mc, objs, count, msec);
    if (msec && n == 0)
        mc->recv_timeouts++;

    return n;
}

msgchan_t *
mc_lookup(const char *name)
{
    msg_chan_t *mc;

    if (name) {
        mc_list_lock();
        TAILQ_FOREACH (mc, &mc_list_head, next) {
            if (!strcmp(name, mc->name)) {
                mc_list_unlock();
                return mc;
            }
        }
        mc_list_unlock();
    }
    return NULL;
}

const char *
mc_name(msgchan_t *_mc)
{
    msg_chan_t *mc = _mc;

    return (mc && mc->cookie == MC_COOKIE) ? mc->name : NULL;
}

int
mc_size(msgchan_t *_mc, int *recv_free_cnt, int *send_free_cnt)
{
    msg_chan_t *mc = _mc;

    if (mc && mc->cookie == MC_COOKIE) {
        int ring1_sz, ring2_sz;

        ring1_sz = rte_ring_free_count(mc->rings[MC_RECV_RING]);
        ring2_sz = rte_ring_free_count(mc->rings[MC_SEND_RING]);

        if (recv_free_cnt)
            *recv_free_cnt = ring1_sz;
        if (send_free_cnt)
            *send_free_cnt = ring2_sz;

        return rte_ring_get_capacity(mc->rings[MC_RECV_RING]);
    }
    return -1;
}

int
mc_info(msgchan_t *_mc, mc_info_t *info)
{
    msg_chan_t *mc = _mc;

    if (mc && info && mc->cookie == MC_COOKIE) {
        info->recv_ring     = mc->rings[MC_RECV_RING];
        info->send_ring     = mc->rings[MC_SEND_RING];
        info->child_count   = mc->nchildren;
        info->send_calls    = mc->send_calls;
        info->send_cnt      = mc->send_cnt;
        info->recv_calls    = mc->recv_calls;
        info->recv_cnt      = mc->recv_cnt;
        info->recv_timeouts = mc->recv_timeouts;
        return 0;
    }

    return -1;
}

void
mc_dump(msgchan_t *_mc)
{
    msg_chan_t *mc = _mc;

    if (mc && mc->cookie == MC_COOKIE) {
        int n = mc_size(_mc, NULL, NULL);
        msg_chan_t *m;

        printf("  %-16s size %d, rings: Recv %p, Send %p Children %d\n", mc->name, n,
               mc->rings[MC_RECV_RING], mc->rings[MC_SEND_RING], mc->nchildren);

        printf("     Send calls %ld count %ld, Recv calls %ld count %ld timeouts %ld\n",
               mc->send_calls, mc->send_cnt, mc->recv_calls, mc->recv_cnt, mc->recv_timeouts);
        if (mc->nchildren) {
            printf("     Children %d: ", mc->nchildren);
            TAILQ_FOREACH (m, &mc->children, next) {
                printf(" %s", m->name);
            }
            printf("\n");
        }
    } else
        TLOG_ERR("MsgChan is invalid\n");
}

void
mc_list(void)
{
    msg_chan_t *mc;

    mc_list_lock();

    printf("** MsgChan **\n");
    TAILQ_FOREACH (mc, &mc_list_head, next)
        mc_dump(mc);

    mc_list_unlock();
}

RTE_INIT_PRIO(mc_constructor, LAST)
{
    TAILQ_INIT(&mc_list_head);

    if (gapi_mutex_create(&mc_list_mutex, PTHREAD_MUTEX_RECURSIVE_NP))
        TLOG_ERR("creating recursive mutex failed\n");
}
