#include "../include/tcp.h"
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

// +OK\r\n
struct SimpleString {
  size_t length;
  std::string data;
};

constexpr const char *CRLF = "\r\n";

void handle_simple_string(tcp::Tcp::Conn &conn) {

  auto &buf = conn.inbuf;

  struct SimpleString simple_string;

  //  "+\r\n"
  if (buf.size() < 3)
    return;

  size_t crlf = buf.find(CRLF);
  if (crlf == std::string::npos)
    return;

  simple_string.data = buf.substr(1, crlf - 1);
  simple_string.length = simple_string.data.size();

  std::cout << simple_string.data << '\n';

  // Consume frame
  buf.erase(0, crlf + 2);
}

void parse_resp(tcp::Tcp::Conn &conn) {

  // Parse it via bulk string for now
  // https://redis.io/docs/latest/develop/reference/protocol-spec/#resp-protocol-description}

  std::string delimiter = "\r\n";

  if (conn.inbuf.size() == 0)
    err(1, "empty inbuf");

  switch (conn.inbuf[0]) {
  case '+':
    // Type simple string
    handle_simple_string(conn);
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
