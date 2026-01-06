#include "../include/tcp.h"

#include <cerrno>
#include <cstring>
#include <err.h>
#include <fcntl.h>
#include <unistd.h>

namespace tcp {

Tcp::Tcp() {
  listen_fd = socket(AF_INET, SOCK_STREAM, 0);
  if (listen_fd == -1)
    err(1, "socket");

  int yes = 1;
  setsockopt(listen_fd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof(yes));

  addr.sin_len = sizeof(addr);
  addr.sin_family = AF_INET;
  addr.sin_port = htons(PORT);
  addr.sin_addr.s_addr = htonl(INADDR_ANY);

  if (bind(listen_fd, (sockaddr *)&addr, sizeof(addr)) == -1)
    err(1, "bind");

  if (listen(listen_fd, 128) == -1)
    err(1, "listen");

  make_nonblocking(listen_fd);

  kqueue_fd = kqueue();
  if (kqueue_fd == -1)
    err(1, "kqueue");

  struct kevent ev;
  EV_SET(&ev, listen_fd, EVFILT_READ, EV_ADD | EV_ENABLE, 0, 0, nullptr);

  if (kevent(kqueue_fd, &ev, 1, nullptr, 0, nullptr) == -1)
    err(1, "kevent register listen_fd");
}

Tcp::~Tcp() {
  (void)close(listen_fd);
  (void)close(kqueue_fd);
}

void Tcp::make_nonblocking(int fd) {
  int flags = fcntl(fd, F_GETFL, 0);
  fcntl(fd, F_SETFL, flags | O_NONBLOCK);
}

std::vector<Tcp::Event> Tcp::poll_events() {
  std::vector<Event> out;

  int nev = kevent(kqueue_fd, nullptr, 0, events, MAX_EVENTS, nullptr);
  if (nev == -1)
    err(1, "kevent poll");

  for (int i = 0; i < nev; i++) {
    auto &kev = events[i];

    if (kev.flags & EV_ERROR) {
      out.push_back({(int)kev.ident, Event::Type::Error});
      continue;
    }

    if (kev.ident == (uintptr_t)listen_fd) {
      out.push_back({listen_fd, Event::Type::Accept});
      continue;
    }

    if (kev.flags & EV_EOF) {
      out.push_back({(int)kev.ident, Event::Type::Close});
      continue;
    }

    if (kev.filter == EVFILT_READ) {
      out.push_back({(int)kev.ident, Event::Type::Read});
    }
  }

  return out;
}

void Tcp::accept_connections() {
  while (true) {
    int client_fd = accept(listen_fd, nullptr, nullptr);
    if (client_fd == -1) {
      if (errno == EAGAIN || errno == EWOULDBLOCK)
        break;
      err(1, "accept");
    }

    make_nonblocking(client_fd);

    struct kevent ev;
    EV_SET(&ev, client_fd, EVFILT_READ, EV_ADD | EV_ENABLE, 0, 0, nullptr);
    kevent(kqueue_fd, &ev, 1, nullptr, 0, nullptr);
  }
}

void Tcp::handle_read(int fd) {
  char buf[1024];

  ssize_t n = read(fd, buf, sizeof(buf));
  if (n > 0) {
    write(fd, buf, n);
  }
  return;
}

} // namespace tcp
