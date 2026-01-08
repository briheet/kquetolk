#ifndef KQUEUE_INCLUDE_DATA_TYPES_H_
#define KQUEUE_INCLUDE_DATA_TYPES_H_

#include "../../include/tcp/tcp.hpp"

namespace DataTypes {

class DataType {
public:
  virtual ~DataType();

  static constexpr const char *CRLF = "\r\n";
  static constexpr size_t CRLF_LEN = 2;

  virtual int read(tcp::Tcp::Conn &conn) = 0;
  virtual int write(tcp::Tcp::Conn &conn, std::string data) = 0;
  virtual int clear(tcp::Tcp::Conn &conn) = 0;
};

} // namespace DataTypes

#endif
