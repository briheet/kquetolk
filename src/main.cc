#include "../include/tcp.h"
#include <cstddef>
#include <cstdio>
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

int main(void) {

  tcp::Tcp server;

  while (true) {

    auto events = server.poll_events();

    for (const auto &ev : events) {

      switch (ev.type) {

      case tcp::Tcp::Event::Type::Accept:
        server.accept_connections();
        break;

      case tcp::Tcp::Event::Type::Close:
        close(ev.fd);
        break;

      case tcp::Tcp::Event::Type::Read:
        server.handle_read(ev.fd);
        break;

      case tcp::Tcp::Event::Type::Error:
        close(ev.fd);
        break;
      }
    }
  }

  return EXIT_SUCCESS;
}
