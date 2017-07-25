#pragma once
#include "libstor-c_types.h"

typedef unsigned long long h;

h* new_client_id(unsigned long long seed);

typedef struct {
    void*       val;
    char*       err;
} result;

typedef struct {
    int            volumes_c;
    char**         volume_ids;
    volume**       volumes;
} volume_map;

volume_map* new_volume_map();

typedef struct {
    int                   services_c;
    char**                service_names;
    volume_map**          volumes;
} service_volume_map;

service_volume_map* new_service_volume_map();
