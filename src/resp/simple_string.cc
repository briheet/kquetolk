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

#include "../../include/resp/simple_string.hpp"
#include "../../include/tcp/tcp.hpp"

namespace SimpleString {

SimpleString::SimpleString() = default;
SimpleString::~SimpleString() = default;

int SimpleString::read(tcp::Tcp::Conn &conn) {
  auto &buf = conn.inbuf;

  if (buf.size() < 3)
    return -1;

  size_t crlf = buf.find(CRLF);
  if (crlf == std::string::npos)
    return -1;

  data = buf.substr(1, crlf - 1);
  length = data.size();

  buf.erase(0, crlf + 2);
  return 0;
};

int SimpleString::write(tcp::Tcp::Conn &conn, std::string data) {

  conn.outbuf.append("+");
  conn.outbuf.append(data);
  conn.outbuf.append("\r\n");
  return 0;
}

int SimpleString::clear(tcp::Tcp::Conn &conn) {
  conn.outbuf.clear();
  return 0;
}

} // namespace SimpleString
