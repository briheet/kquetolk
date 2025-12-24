#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <err.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/event.h>
#include <unistd.h>

class Kqueue {

  int kq;
  struct kevent event;
  struct kevent triggeredEvent;

public:
  Kqueue(void) {
    // Create kqueue
    this->kq = kqueue();
    if (kq == -1) {
      err(EXIT_FAILURE, "kqueue failure");
    }
  }

  void set_event(uintptr_t ident, int16_t filter, uint16_t flags,
                 uint32_t fflags, intptr_t data, void *udata) {

    EV_SET(&event, ident, filter, flags, fflags, data, udata);
  }

  int call_kevent(int nchanges, int nevents, const struct timespec *timeout) {

    return kevent(kq, &event, nchanges, &triggeredEvent, nevents, timeout);
  }

  const struct kevent &get_triggered_event() { return triggeredEvent; }

  ~Kqueue() { close(kq); }
};

int main(void) {

  Kqueue kq;

  // Set event
  kq.set_event(1, EVFILT_TIMER, EV_ADD | EV_ENABLE, 0, 500, NULL);

  int count = 5;

  for (;;) {

    // Sleep untill something happens
    int ret = kq.call_kevent(1, 1, NULL);
    if (ret == -1) {
      err(EXIT_FAILURE, "kevent wait");
    } else if (ret > 0) {
      const auto &ev = kq.get_triggered_event();
      if (ev.flags & EV_ERROR) {
        errx(EXIT_FAILURE, "Event error: %s", strerror(ev.data));
      }

      pid_t pid = fork();

      if (pid < 0) {
        err(EXIT_FAILURE, "fork()");
      } else if (pid == 0) {
        if (execlp("date", "date", (char *)0) < 0) {
          err(EXIT_FAILURE, "execlp()");
        }
      }
    }

    count = count - 1;

    if (count == 0)
      break;
  }

  return EXIT_SUCCESS;
}
