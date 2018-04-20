#include <algorithm>

#include <cstring>
#include <cerrno>
#include <cstdio>

#include <sys/types.h>
#include <sys/socket.h>

#include "ChunkedDataSender.h"

ArraySender::ArraySender(const char *array_to_send, size_t length) {
	this->array = new char[length];
	std::copy(array_to_send, array_to_send+length, this->array);
	this->array_length = length;
	this->curr_loc = 0;
}

ssize_t ArraySender::send_next_chunk(int sock_fd) {
	size_t num_bytes_remaining = array_length - curr_loc;
	size_t bytes_in_chunk = std::min(num_bytes_remaining, CHUNK_SIZE);
	if (bytes_in_chunk > 0) {
		char chunk[CHUNK_SIZE];
		memcpy(chunk, array+curr_loc, bytes_in_chunk);
		ssize_t num_bytes_sent = send(sock_fd, chunk, bytes_in_chunk, 0);

		if (num_bytes_sent < 0 && errno != EAGAIN) {
			perror("send_next_chunk send");
			exit(EXIT_FAILURE);
		}
		else if (num_bytes_sent > 0) {
			curr_loc += num_bytes_sent;
		}

		return num_bytes_sent;
	}
	else {
		return 0;
	}
}
