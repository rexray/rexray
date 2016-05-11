#include <stdio.h>
#include <stdlib.h>
#include <inttypes.h>
#include "libstor-c.h"

int main(int argc, char** argv) {

    if (argc < 2) {
        printf("usage: libstor-c CONFIG\n");
        return 1;
    }

    result c = new_client(argv[1]);
    if (c.err) {
        printf("libstor-c: error: %s\n", c.err);
        return 1;
    }
    h client_id = *((h*)c.val);

    result v = volumes(client_id, 0);
    if (v.err) {
        printf("libstor-c: error: %s\n", v.err);
        close(client_id);
        return 1;
    }
    service_volume_map svm = *((service_volume_map*)v.val);

    for (int x = 0; x < svm.services_c; x++) {

        printf("service: %s\n", svm.service_names[x]);

        volume_map vm = *svm.volumes[x];

        for (int y = 0; y < vm.volumes_c; y++) {
            volume vol = *vm.volumes[y];
            printf("\n");
            printf("  - volume: %s\n", vm.volume_ids[y]);
            printf("            name:   %s\n", vol.name);
            printf("            iops:   %" PRId64 "\n", vol.iops);
            printf("            size:   %" PRId64 "\n", vol.size);
            printf("            type:   %s\n", vol.volume_type);
            printf("            zone:   %s\n", vol.availability_zone);
            printf("            status: %s\n", vol.status);
            printf("            netnam: %s\n", vol.network_name);
        }

        printf("\n");
    }

    close(client_id);
	return 0;

}
