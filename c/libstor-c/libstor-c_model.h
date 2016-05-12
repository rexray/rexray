#pragma once
#include <stdint.h>
#include "libstor-c_types.h"

// type of storage a driver provides.
enum STORAGE_TYPES {
    BLOCK,
    NAS,
    OBJECT
};

// identifies a host to a remote storage platform
typedef struct {
    char*       id;
} instance_id;

instance_id* new_instance_id();

// provides information about a client to a storage platform
typedef struct {
    instance_id*     instance_id;
    char*            name;
    char*            provider_name; // name of the provider that owns the object
    char*            region;        // region from which the object originates
} instance;

instance* new_instance();

// provides information about an object attached to a volume
typedef struct {
    char*            volume_id;
    instance_id*     instance_id;
    char*            device_name;
    char*            mount_point;
    char*            status;
} volume_attachment;

volume_attachment* new_volume_attachment();

// a storage volume
typedef struct {
    char*      id;
    char*      name;
    int64_t    iops;
    int64_t    size;                // the volume's size in bytes
    char*      status;
    char*      volume_type;
    char*      availability_zone;  // zone for which the volume is available
    char*      network_name;       // name the device is known by in order to
                                    // discover locally

    int                 attachments_c;    // number of attachments
    volume_attachment** attachments;     // the volume's attachments
} volume;

volume* new_volume();
