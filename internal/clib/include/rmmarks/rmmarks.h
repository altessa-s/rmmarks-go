/* Copyright 2026 ALTESSA SOLUTIONS INC.
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file. */
#ifndef RMMARKS_H
#define RMMARKS_H
#include "rmmarks/text.h"
#include "rmmarks/types.h"
#define RMMARKS_VERSION_MAJOR 1
#define RMMARKS_VERSION_MINOR 0
#define RMMARKS_VERSION_PATCH 1
#define RMMARKS_STRINGIFY_(x) #x
#define RMMARKS_STRINGIFY(x) RMMARKS_STRINGIFY_(x)
#define RMMARKS_VERSION_STRING                                                                     \
    RMMARKS_STRINGIFY(RMMARKS_VERSION_MAJOR)                                                       \
    "." RMMARKS_STRINGIFY(RMMARKS_VERSION_MINOR) "." RMMARKS_STRINGIFY(RMMARKS_VERSION_PATCH)
#endif
