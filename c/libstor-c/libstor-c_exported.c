#include <stdlib.h>
#include "libstor-c_types.h"

h* new_client_id(unsigned long long val) {
    h* client_id = (h*)malloc(sizeof(h));
    *client_id = val;
    return client_id;
}

volume_map* new_volume_map() {
    return (volume_map*)malloc(sizeof(volume_map));
}

service_volume_map* new_service_volume_map() {
    return (service_volume_map*)malloc(sizeof(service_volume_map));
}
