#ifndef KQUETOLK_INCLUDE_TCP_H_
#define KQUETOLK_INCLUDE_TCP_H_

#include <cstdint>
#include <netinet/in.h>
#include <sys/event.h>
#include <unordered_map>
#include <vector>

namespace tcp {

class Tcp {
public:
  struct Event {
    enum class Type { Accept, Read, Close, Error };

    uintptr_t fd;
    Type type;
  };

  struct Conn {
    int fd;
    std::string inbuf;
    std::string outbuf;
  };

  Tcp();
  ~Tcp();

  std::vector<Event> poll_events();
  void accept_connections(const Event &);
  void handle_read(const Event &ev);

  std::unordered_map<int, Conn> connections;

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
