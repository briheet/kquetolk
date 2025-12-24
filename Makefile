CXX = clang++
CXXFLAGS = -std=c++23 -Wall -W -Wextra -pedantic

run:
	@./bin/main

build:
	$(CXX) $(CXXFLAGS) src/main.cc -o ./bin/main

flame:
	@/Users/briheet/.cargo/bin/flamegraph -- bin/main
