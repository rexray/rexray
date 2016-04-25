// +build linux darwin

package semaphore

// #include <fcntl.h>           /* For O_* constants */
// #include <sys/stat.h>        /* For mode constants */
// #include <semaphore.h>
// #include <string.h>
// #include <stdlib.h>
// #include <errno.h>
// #include <stdio.h>
/*
int _errno() {
    return errno;
}

typedef struct {
    sem_t*      val;
    int         err;
} sem_tt;

sem_tt* _sem_open(char* name, int flags, mode_t perm, unsigned int val) {
    sem_tt* r = (sem_tt*)malloc(sizeof(sem_tt));
    sem_t* sem = sem_open((const char*)name, flags, perm, val);
	if (sem == SEM_FAILED) r->err = errno;
    else { r->err = 0; r->val = sem; }
    return r;
}

int _sem_close(void* sem) {
    return sem_close(((sem_tt*)sem)->val) == 0 ? 0 : errno;
}

int _sem_wait(void* sem) {
	while (sem_wait(((sem_tt*)sem)->val))
		if (errno == EINTR) errno = 0;
		else return errno;
	return 0;
}

int _sem_trywait(void* sem) {
    while (sem_trywait(((sem_tt*)sem)->val))
		if (errno == EINTR) errno = 0;
		else return errno;
	return 0;
}

int _sem_post(void* sem) {
    return sem_post(((sem_tt*)sem)->val) == 0 ? 0 : errno;
}

int _sem_unlink(char* name) {
    return sem_unlink((const char*) name) == 0 ? 0 : errno;
}
*/
import "C"

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/akutz/goof"
)

type semaphore struct {
	name  string
	sema  unsafe.Pointer
	count int64
}

func open(
	name string, excl bool, perm os.FileMode, val int) (Semaphore, error) {

	if !strings.HasPrefix(name, `/`) {
		name = fmt.Sprintf("/%s", name)
	}
	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))

	flags := C.O_CREAT
	if excl {
		flags = flags | C.O_EXCL
	}

	sema := C._sem_open(cs, C.int(flags), C.mode_t(perm), C.uint(val))
	if sema.err != 0 {
		return nil, goof.WithFields(goof.Fields{
			"name":  name,
			"error": sema.err,
		}, "error opening semaphore")
	}

	return &semaphore{
		name:  name,
		sema:  unsafe.Pointer(sema),
		count: int64(val),
	}, nil
}

func (s *semaphore) Name() string {
	return s.name
}

func (s *semaphore) Close() error {
	err := C._sem_close(s.sema)
	if err == 0 {
		return nil
	}
	C.free(s.sema)
	return goof.WithFields(goof.Fields{
		"name":  s.name,
		"error": int(err),
	}, "error closing semaphore")
}

func (s *semaphore) Signal() error {
	err := C._sem_post(s.sema)
	if err == 0 {
		atomic.AddInt64(&s.count, 1)
		return nil
	}
	return goof.WithFields(goof.Fields{
		"name":  s.name,
		"error": int(err),
	}, "error unlocking semaphore")
}

func (s *semaphore) Value() (int, error) {
	return s.value()
}

func (s *semaphore) Wait() error {
	err := C._sem_wait(s.sema)
	if err == 0 {
		atomic.AddInt64(&s.count, -1)
		return nil
	}
	return goof.WithFields(goof.Fields{
		"name":  s.name,
		"error": int(err),
	}, "error waiting on semaphore")
}

func (s *semaphore) TryWait() (bool, error) {
	err := C._sem_trywait(s.sema)
	if err == 0 {
		atomic.AddInt64(&s.count, -1)
		return false, nil
	} else if err == C.EAGAIN {
		return true, nil
	}
	return false, goof.WithFields(goof.Fields{
		"name":  s.name,
		"error": int(err),
	}, "error trying wait on semaphore")
}

func (s *semaphore) TimedWait(t time.Duration) error {
	return s.timedWait(t)
}

func unlink(name string) error {
	if !strings.HasPrefix(name, `/`) {
		name = fmt.Sprintf("/%s", name)
	}
	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))
	err := C._sem_unlink(cs)
	if err == 0 {
		return nil
	}
	return goof.WithFields(goof.Fields{
		"name":  name,
		"error": int(err),
	}, "error unlinking semaphore")
}
