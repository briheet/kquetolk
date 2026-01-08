#ifndef KQUEUE_INCLUDE_RESP_H_
#define KQUEUE_INCLUDE_RESP_H_

#include "../../include/data_types/data_types.h"
#include "../../include/tcp/tcp.hpp"

namespace Resp {

class resp {

public:
  enum class Type {
    // DataTypes::dataTypes::Type::Simple_string,
  };

  int parse(std::string &buf, resp::Type);
};

} // namespace Resp

#endif
