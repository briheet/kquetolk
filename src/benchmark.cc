#include <benchmark/benchmark.h>

static void BM_Addition(benchmark::State &state) {
  for (auto _ : state) {
    int x = 0;
    for (int i = 0; i < 1000; ++i)
      x += i;
    benchmark::DoNotOptimize(x); // prevents "unused variable" warning
  }
}
BENCHMARK(BM_Addition);

BENCHMARK_MAIN();
