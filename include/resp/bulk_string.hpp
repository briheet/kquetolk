#ifndef KQUEUE_INCLUDE_RESP_BULK_STRING_H_
#define KQUEUE_INCLUDE_RESP_BULK_STRING_H_

#include "../../include/data_types/data_types.hpp"

namespace BulkString {

class BulkString : public DataTypes::DataType {

public:
  BulkString();
  ~BulkString();

  int read(tcp::Tcp::Conn &conn) override;
  int write(tcp::Tcp::Conn &conn, std::string data) override;
  int clear(tcp::Tcp::Conn &conn) override;

  size_t length = 0;
  std::string data = "";
};

} // namespace BulkString

#endif
