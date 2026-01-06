#ifndef KQUETOLK_INCLUDE_TCP_H_
#define KQUETOLK_INCLUDE_TCP_H_

#include <cstdint>
#include <netinet/in.h>
#include <sys/event.h>
#include <vector>

namespace tcp {

class Tcp {
public:
  struct Event {
    enum class Type { Accept, Read, Close, Error };

    int fd;
    Type type;
  };

  Tcp();
  ~Tcp();

  std::vector<Event> poll_events();
  void accept_connections();
  void handle_read(int fd);

private:
  static constexpr std::uint32_t MAX_EVENTS = 128;
  static constexpr std::uint16_t PORT = 6379;

  int listen_fd;
  int kqueue_fd;

  sockaddr_in addr{};
  struct kevent events[MAX_EVENTS];

  void make_nonblocking(int fd);
};

} // namespace tcp

#endif
