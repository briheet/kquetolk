#include <cstddef>
#include <cstdlib>
#include <cstring>
#include <err.h>
#include <fcntl.h>
#include <iostream>
#include <netinet/in.h>
#include <sys/_types/_ssize_t.h>
#include <sys/event.h>
#include <sys/socket.h>
#include <unistd.h>

// Plan: Create a socket, give it a address and port, listen for connections,
// accept client and send and recieve data
constexpr int MAX_EVENTS = 1024;

int main(void) {
  std::cout << "Hi" << "\n";

  int tcp_socket_fd = socket(AF_INET, SOCK_STREAM, 0);
  if (tcp_socket_fd == -1) {
    err(EXIT_FAILURE, "tcp socket init error");
  }

  struct sockaddr_in tcp_socket_addr;
  std::memset(&tcp_socket_addr, 0, sizeof(tcp_socket_addr));

  tcp_socket_addr.sin_len = sizeof(tcp_socket_addr); // BSD/Macos stuff
  tcp_socket_addr.sin_family = AF_INET;
  tcp_socket_addr.sin_port = htons(6379);
  tcp_socket_addr.sin_addr.s_addr = htonl(INADDR_ANY);

  int yes = 1;
  setsockopt(tcp_socket_fd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof(yes));

  // Bind
  int ret =
      bind(tcp_socket_fd, reinterpret_cast<struct sockaddr *>(&tcp_socket_addr),
           sizeof(tcp_socket_addr));
  if (ret == -1) {
    err(EXIT_FAILURE, "bind issue\n");
  }

  ret = listen(tcp_socket_fd, 50);
  if (ret == -1) {
    err(EXIT_FAILURE, "listen error\n");
  }

  int flags = fcntl(tcp_socket_fd, F_GETFL, 0);
  fcntl(tcp_socket_fd, F_SETFL, flags | O_NONBLOCK);

  // Create a kqueue to watch for events
  int kq = kqueue();
  if (kq == -1) {
    err(EXIT_FAILURE, "kqueue open failure\n");
  }

  struct kevent tcp_events{};
  EV_SET(&tcp_events, tcp_socket_fd, EVFILT_READ, EV_ADD | EV_ENABLE, 0, 0,
         nullptr);

  // Attach event in kqueue
  ret = kevent(kq, &tcp_events, 1, nullptr, 0, nullptr);
  if (ret == -1) {
    err(EXIT_FAILURE, "unable to attach kevent in kqueue\n");
  }

  struct kevent events[MAX_EVENTS];

  for (;;) {

    int nev = kevent(kq, nullptr, 0, events, MAX_EVENTS, nullptr);
    if (nev == -1) {
      err(EXIT_FAILURE, "kevent_wait\n");
    }

    for (int i = 0; i < nev; i++) {

      int event_fd = static_cast<int>(events[i].ident);
      if (event_fd == tcp_socket_fd) {

        // New client conn
        for (;;) {
          struct sockaddr client_addr{};
          socklen_t addr_len = sizeof(client_addr);

          int client_fd = accept(
              tcp_socket_fd, reinterpret_cast<struct sockaddr *>(&client_addr),
              &addr_len);

          if (client_fd == -1) {
            if (errno == EAGAIN || errno == EWOULDBLOCK)
              break; // accept queue drained
            err(EXIT_FAILURE, "accept issue\n");
          }

          int flags = fcntl(client_fd, F_GETFL, 0);
          if (flags == -1)
            err(EXIT_FAILURE, "fcntl(F_GETFL)");
          if (fcntl(client_fd, F_SETFL, flags | O_NONBLOCK) == -1)
            err(EXIT_FAILURE, "fcntl(F_SETFL)");

          struct kevent cev{};
          EV_SET(&cev, client_fd, EVFILT_READ, EV_ADD | EV_ENABLE, 0, 0,
                 nullptr);

          kevent(kq, &cev, 1, nullptr, 0, nullptr);
        }
      } else if (events[i].filter == EVFILT_READ) {
        // Client read

        char buf[4096];
        ssize_t size = recv(event_fd, buf, sizeof(buf), 0);

        // Debug
        std::cout.write(buf, size);
        std::cout << std::endl;

        if (size == 0) {
          close(event_fd);
          continue;
        }

        if (size < 0) {

          if (errno != EAGAIN && errno != EWOULDBLOCK)
            close(event_fd);
          continue;
        }

        char reply[] = "gotchya message\n";
        send(event_fd, reply, sizeof(reply) - 1, 0);
      }
    }
  }

  (void)close(tcp_socket_fd);
  (void)close(kq);

  return EXIT_SUCCESS;
}
