#include <stdio.h>
#include <semaphore.h>
#include <errno.h>
#include <stdlib.h>

int main(int argc, char** argv) {
	if (argc < 2) {
		printf("unlink: %s <semaphore_name>\n", argv[0]);
		return 1;
	}
	return sem_unlink(argv[1]) ? errno : 0;
}
