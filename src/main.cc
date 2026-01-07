#include "../include/tcp.h"
#include <cstddef>
#include <cstdlib>
#include <cstring>
#include <err.h>
#include <fcntl.h>
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
        server.accept_connections(ev);
        break;

      case tcp::Tcp::Event::Type::Close:
        server.connections.erase(ev.fd);
        close(ev.fd);
        break;

      case tcp::Tcp::Event::Type::Read:
        server.handle_read(ev);
        parse_resp(server.connections[ev.fd]);
        break;

      case tcp::Tcp::Event::Type::Error:
        close(ev.fd);
        server.connections.erase(ev.fd);
        break;
      }
    }
  }

  return EXIT_SUCCESS;
}

void parse_resp(tcp::Tcp::Conn &conn) {
  size_t pos = 0;

  while (pos < conn.inbuf.size()) {
    size_t cmd_len =
        try_parse_resp(conn.inbuf.data() + pos, conn.inbuf.size() - pos);
    if (cmd_len == 0)
      break; // incomplete

    execute_resp_command(conn.inbuf.data() + pos, cmd_len, conn.outbuf);
    pos += cmd_len;
  }

  conn.inbuf.erase(0, pos);
}
