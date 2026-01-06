CXX = clang++
CXXFLAGS = -std=c++23 -Wall -W -Wextra -pedantic
SRC_INCLUDE = ./include
SRC_LIB = ./src
BENCH_INCLUDE = ../benchmark/include
BENCH_LIB = ../benchmark/build/src
BIN = ./bin

runapp:
	$(BIN)/main

buildapp:
	$(CXX) $(CXXFLAGS) -I$(SRC_INCLUDE) src/main.cc src/tcp.cc -o ./bin/main

flame:
	/Users/briheet/.cargo/bin/flamegraph -- $(BIN)/main

buildbench:
	$(CXX) $(CXXFLAGS) -I$(BENCH_INCLUDE) src/benchmark.cc -L$(BENCH_LIB) -lbenchmark -lpthread -o $(BIN)/mybenchmark

runbench:
	$(BIN)/mybenchmark
