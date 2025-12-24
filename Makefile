run:
	@./bin/main

build:
	@clang++ -std=c++23 src/main.cc -o ./bin/main -Wall -W -Wextra -ansi -pedantic
