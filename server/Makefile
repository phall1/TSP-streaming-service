# This makefile compiles the server example.
# You should modify it to also compile a program called jukebox-server.
# jukebox-server will be the server implementation of your Jukebox.

CXX = g++
CXXFLAGS=-Wall -Wextra -g -O1 -std=c++17 -lboost_system -lboost_filesystem

SRC_FILES = server-example.cpp
TARGETS = server-example

all: $(TARGETS)

server-example: $(SRC_FILES)
	$(CXX) $(CXXFLAGS) -o $@ $^

clean:
	rm -f $(TARGETS)
