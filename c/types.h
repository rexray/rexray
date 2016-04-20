#include <stdint.h>

// type of storage a driver provides.
enum STORAGE_TYPES {
    BLOCK,
    NAS,
    OBJECT
};

// a client handle
typedef unsigned long long hc;

// an error
typedef struct {
    char      *msg;
} error;

// identifies a host to a remote storage platform
typedef struct {
    char       *id;
    char       *metadata;   // extra information about the instance id,
                            // marshaled as a JSON string
} instance_id;

// provides information about a client to a storage platform
typedef struct {
    instance_id     *instance_id;
    char            *name;
    char            *provider_name; // name of the provider that owns the object
    char            *region;        // region from which the object originates
} instance;

// provides information about an object attached to a volume
typedef struct {
    char            *volume_id;
    instance_id     *instance_id;
    char            *device_name;
    char            *mount_point;
    char            *status;
} volume_attachment;

// a storage volume
typedef struct {
    char       *id;
    char       *name;
    int64_t    iops;
    int64_t    size;                // the volume's size in bytes
    char       *status;
    char       *volume_type;
    char       *availability_zone;  // zone for which the volume is available
    char       *network_name;       // name the device is known by in order to
                                    // discover locally

    int               attachments_c;    // number of attachments
    volume_attachment *attachments;     // the volume's attachments
} volume;
