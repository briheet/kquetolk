#include <cstdio>
#include <cstdlib>
#include <err.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/event.h>
#include <unistd.h>

int main(void) {

  // Event and triggered event
  struct kevent event;
  struct kevent triggeredEvent;

  // kqueue, ret vals
  int kq, ret;

  // Create Kqueue
  kq = kqueue();
  if (kq == -1)
    err(EXIT_FAILURE, "kqueue() failed");

  // Init a event in kqueue
  EV_SET(&event, 1, EVFILT_TIMER, EV_ADD | EV_ENABLE, 0, 5000, NULL);

  // loop my nig
  for (;;) {

    // Sleep untill something happens
    ret = kevent(kq, &event, 1, &triggeredEvent, 1, NULL);

    if (ret == -1) {
      err(EXIT_FAILURE, "kevent wait");
    } else if (ret > 0) {
      if (triggeredEvent.flags & EV_ERROR) {
        errx(EXIT_FAILURE, "Event error: %s", strerror(event.data));
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
  }

  // Destroy
  (void)close(kq);
  return EXIT_SUCCESS;
}
