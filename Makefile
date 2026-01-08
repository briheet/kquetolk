CXX = clang++
CXXFLAGS = -std=c++23 -g -Wall -W -Wextra -pedantic -fno-omit-frame-pointer
SRC_INCLUDE = ./include
SRC_LIB = ./src
BENCH_INCLUDE = ../benchmark/include
BENCH_LIB = ../benchmark/build/src
BIN = ./bin

runapp:
	$(BIN)/main

buildapp:
	$(CXX) $(CXXFLAGS) -I$(SRC_INCLUDE) src/main.cc src/tcp/tcp.cc src/resp/simple_string.cc src/resp/bulk_string.cc src/data_types/data_types.cc -o ./bin/main

flame:
	/Users/briheet/.cargo/bin/flamegraph -- $(BIN)/main

buildbench:
	$(CXX) $(CXXFLAGS) -I$(BENCH_INCLUDE) src/benchmark.cc -L$(BENCH_LIB) -lbenchmark -lpthread -o $(BIN)/mybenchmark

runbench:
	$(BIN)/mybenchmark
