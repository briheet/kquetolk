#include <cstddef>
#include <cstdlib>
#include <cstring>
#include <err.h>
#include <fcntl.h>
#include <netinet/in.h>
#include <string>
#include <sys/_types/_ssize_t.h>
#include <sys/event.h>
#include <sys/socket.h>
#include <unistd.h>

#include "../include/resp/bulk_string.hpp"
#include "../include/resp/simple_string.hpp"
#include "../include/tcp/tcp.hpp"

void parse_resp(tcp::Tcp::Conn &conn) {

  if (conn.inbuf.size() == 0)
    err(1, "empty inbuf");

  switch (conn.inbuf[0]) {
  case '+': {
    SimpleString::SimpleString value;
    if (value.read(conn) == 0) {
      value.write(conn, value.data);
    }
    break;
  }
  case '$': {
    BulkString::BulkString value;
    if (value.read(conn) == 0) {
      value.write(conn, value.data);
    }
    break;
  }
  case '*': {
  }
  }
}

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
        server.connections.erase(ev.fd);
        close(ev.fd);
        break;

      case tcp::Tcp::Event::Type::Read:
        server.handle_read(ev);
        parse_resp(server.connections[ev.fd]);
        server.enable_write(ev.fd);
        break;

      case tcp::Tcp::Event::Type::Write:
        server.handle_write(ev);
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
