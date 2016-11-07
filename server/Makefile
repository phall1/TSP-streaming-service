# This makefile compiles the server example.
# You should modify it to also compile a program called jukebox-server.
# jukebox-server will be the server implementation of your Jukebox.

CC = gcc
CFLAGS = -g -O1 -Wall -Wextra

TARGETS = server-example

all: $(TARGETS)

server-example: server-example.c
	$(CC) $(CFLAGS) -o $@ $^

clean:
	rm -f $(TARGETS)
