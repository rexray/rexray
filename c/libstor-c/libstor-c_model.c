#include <stdlib.h>
#include "libstor-c_types.h"

volume* new_volume() {
    return (volume*)malloc(sizeof(volume));
}

volume_attachment* new_volume_attachment() {
    return (volume_attachment*)malloc(sizeof(volume_attachment));
}

instance* new_instance() {
    return (instance*)malloc(sizeof(instance));
}

instance_id* new_instance_id() {
    return (instance_id*)malloc(sizeof(instance_id));
}
