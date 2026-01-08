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
#include <vector>

#include "../../include/resp/bulk_string.hpp"
#include "../../include/tcp/tcp.hpp"

namespace BulkString {

BulkString::BulkString() = default;
BulkString::~BulkString() = default;

int BulkString::read(tcp::Tcp::Conn &conn) {
  auto &buf = conn.inbuf;

  if (buf.size() < 3)
    return -1;

  std::vector<std::string> tokens;
  std::string token;
  size_t pos = 0;

  while ((pos = buf.find(CRLF)) != std::string::npos) {
    token = buf.substr(0, pos);
    tokens.emplace_back(token);
    buf.erase(0, pos + CRLF_LEN);
  }

  data = tokens[1];
  length = tokens[1].size();
  return 0;
}

int BulkString::write(tcp::Tcp::Conn &conn, std::string data) {

  conn.outbuf.append("$");
  conn.outbuf.append(data);
  conn.outbuf.append("\r\n");
  return 0;
}

int BulkString::clear(tcp::Tcp::Conn &conn) {

  conn.outbuf.clear();
  return 0;
}

} // namespace BulkString
