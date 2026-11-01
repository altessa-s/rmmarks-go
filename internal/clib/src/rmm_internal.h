/* Copyright 2026 ALTESSA SOLUTIONS INC.
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file. */
#ifndef RMM_INTERNAL_H
#define RMM_INTERNAL_H
#include "rmmarks/rmmarks.h"
#include <math.h>
#include <stdlib.h>
#include <string.h>

typedef struct rmm_arena rmm_arena;
rmm_arena *rmm_arena_create(void);
void rmm_arena_destroy(rmm_arena *arena);
void *rmm_arena_alloc(rmm_arena *arena, size_t size);
void *rmm_arena_calloc(rmm_arena *arena, size_t count, size_t size);
char *rmm_arena_strndup(rmm_arena *arena, const char *bytes, size_t length);
#ifdef RMMARKS_TESTING
void rmm_test_fail_after(size_t count);
bool rmm_test_did_fail(void);
#endif
bool rmm_array_valid(const void *data, size_t count, size_t element);
bool rmm_box_valid(rmm_page_box box);
rmm_page_box rmm_box_union(rmm_page_box a, rmm_page_box b);
rmm_status rmm_finish(rmm_error *error, rmm_status status);
rmm_status rmm_validate_lines(const rmm_line *lines, size_t count, size_t *pieces);
bool rmm_utf8_valid(rmm_str text);
size_t rmm_utf8_next(rmm_str text, size_t offset, uint32_t *scalar);
rmm_str rmm_trim(rmm_str text);
rmm_status rmm_text_equal(rmm_arena *arena, rmm_str a, rmm_str b, bool *equal);

#endif
