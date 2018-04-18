# This makefile compiles the server example.
# You should modify it to also compile a program called jukebox-server.
# jukebox-server will be the server implementation of your Jukebox.

CXX = g++
CXXFLAGS = -g -O1 -Wall -Wextra -std=c++11

SRC_FILES = server-example.cpp
TARGETS = server-example

all: $(TARGETS)

server-example: $(SRC_FILES)
	$(CXX) $(CXXFLAGS) -o $@ $^

clean:
	rm -f $(TARGETS)
