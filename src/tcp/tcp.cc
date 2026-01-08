
#include <cerrno>
#include <cstring>
#include <err.h>
#include <fcntl.h>
#include <sys/event.h>
#include <unistd.h>

#include "../../include/tcp/tcp.hpp"

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
      out.emplace_back(Event{
          .fd = kev.ident,
          .type = Event::Type::Error,
      });
      continue;
    }

    if (kev.ident == (uintptr_t)listen_fd) {
      out.emplace_back(Event{
          .fd = kev.ident,
          .type = Event::Type::Accept,
      });
      continue;
    }

    if (kev.flags & EV_EOF) {
      out.emplace_back(Event{
          .fd = kev.ident,
          .type = Event::Type::Close,
      });
      continue;
    }

    if (kev.filter == EVFILT_READ) {
      out.emplace_back(Event{
          .fd = kev.ident,
          .type = Event::Type::Read,
      });
      continue;
    }

    if (kev.filter == EVFILT_WRITE) {
      out.emplace_back(Event{
          .fd = kev.ident,
          .type = Event::Type::Write,
      });
      continue;
    }
  }

  return out;
}

void Tcp::accept_connections() {
  for (;;) {
    int client_fd = accept(listen_fd, nullptr, nullptr);
    if (client_fd == -1) {
      if (errno == EAGAIN || errno == EWOULDBLOCK)
        break;
      err(1, "accept");
    }

    make_nonblocking(client_fd);

    // Add to map
    Conn c;
    c.fd = client_fd;
    connections[client_fd] = std::move(c);

    struct kevent ev;
    EV_SET(&ev, client_fd, EVFILT_READ, EV_ADD | EV_ENABLE, 0, 0, nullptr);
    if (kevent(kqueue_fd, &ev, 1, nullptr, 0, nullptr) == -1)
      err(1, "kevent register client");
  }
}

void Tcp::handle_read(const Event &ev) {

  auto it = connections.find(ev.fd);
  if (it == connections.end()) {
    close(ev.fd);
    return;
  }

  Conn &c = it->second;
  char buf[1024];

  for (;;) {
    ssize_t n = read(ev.fd, buf, sizeof(buf));

    if (n > 0) {
      c.inbuf.append(buf, n);
    } else if (n == 0) {
      close(ev.fd);
      connections.erase(it);
      break;
    } else {
      if (errno == EAGAIN || errno == EWOULDBLOCK)
        break;
      close(ev.fd);
      connections.erase(it);
      break;
    }
  }
}

void Tcp::enable_write(int fd) {
  struct kevent ev;
  EV_SET(&ev, fd, EVFILT_WRITE, EV_ADD | EV_ENABLE, 0, 0, nullptr);
  kevent(kqueue_fd, &ev, 1, nullptr, 0, nullptr);
}

void Tcp::handle_write(const Event &ev) {

  auto it = connections.find(ev.fd);

  if (it == connections.end())
    return;

  Conn &conn = it->second;

  if (conn.outbuf.empty())
    return;

  ssize_t n = write(conn.fd, conn.outbuf.data(), conn.outbuf.size());

  if (n <= 0) {
    if (errno == EAGAIN || errno == EWOULDBLOCK)
      return;

    close(ev.fd);
    connections.erase(it);
    return;
  }

  conn.outbuf.erase(0, n);

  // Disable write notifications if done
  if (conn.outbuf.empty()) {
    struct kevent kev;
    EV_SET(&kev, ev.fd, EVFILT_WRITE, EV_DELETE, 0, 0, nullptr);
    kevent(kqueue_fd, &kev, 1, nullptr, 0, nullptr);
  }
}

} // namespace tcp
