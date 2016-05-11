#include <stdio.h>
#include <semaphore.h>
#include <errno.h>
#include <fcntl.h>

int main(int argc, char** argv) {
	if (argc < 2) {
		printf("signal: %s <semaphore_name>\n", argv[0]);
		return 1;
	}
	sem_t* sem = sem_open(argv[1], O_CREAT, 0644, 1);
	if (sem == SEM_FAILED) {
		return errno;
	}
	return sem_post(sem) ? errno : 0;
}
