package semaphore

// #include <fcntl.h>           /* For O_* constants */
// #include <sys/stat.h>        /* For mode constants */
// #include <semaphore.h>
// #include <string.h>
// #include <stdlib.h>
// #include <errno.h>
/*
typedef struct {
    sem_t*      val;
    int         err;
} sem_tt;

int _sem_timedwait(void* sem, time_t sec, long nsec) {
	struct timespec* ts = (struct timespec*)malloc(sizeof(struct timespec));
    ts->tv_sec = sec;
    ts->tv_nsec = nsec;
	const struct timespec* cts = (const struct timespec*) ts;
	while (sem_timedwait(((sem_tt*)sem)->val, cts))
		if (errno == EINTR) errno = 0;
		else return errno;
	return 0;
}

typedef struct {
	int     ret;
	int     val;
} getvalue_result;

getvalue_result* _sem_getvalue(void* sem) {
	getvalue_result* r = (getvalue_result*)malloc(sizeof(getvalue_result));
	r->ret = sem_getvalue(((sem_tt*)sem)->val, &(r->val)) == 0 ? 0 : errno;
	return r;
}
*/
import "C"
import (
	"time"

	"github.com/akutz/goof"
)

func (s *semaphore) timedWait(t *time.Time) error {
	err := C._sem_timedwait(s.sema, C.time_t(t.Unix()), C.long(t.UnixNano()))
	if err == 0 {
		return nil
	}
	return goof.WithFields(goof.Fields{
		"name":  s.name,
		"error": int(err),
	}, "error timed waiting on semaphore")
}

func (s *semaphore) value() (int, error) {
	r := C._sem_getvalue(s.sema)
	if r.ret == 0 {
		return int(r.val), nil
	}
	return 0, goof.WithFields(goof.Fields{
		"name":  s.name,
		"error": int(r.ret),
	}, "error getting semaphore value")
}
