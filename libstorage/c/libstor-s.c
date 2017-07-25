#include <stdio.h>
#include <stdlib.h>
#include "libstor-s.h"

int main(int argc, char** argv) {
    if (argc < 1) {
        printf("usage: libstor-s DRIVER\n");
        return 1;
    }

    closeOnAbort();
    serve("", argc-1, argv[1]);
	return 0;
}
