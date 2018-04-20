#ifndef CHUNKEDDATASENDER_H
#define CHUNKEDDATASENDER_H

#include <cstddef>

const size_t CHUNK_SIZE = 4096;

class ChunkedDataSender {
  public:
	virtual ~ChunkedDataSender() {}

	virtual ssize_t send_next_chunk(int sock_fd) = 0;
};

class ArraySender : public virtual ChunkedDataSender {
  private:
	char *array;
	size_t array_length;
	size_t curr_loc;

  public:
	ArraySender(const char *array_to_send, size_t length);

	~ArraySender() {
		delete[] array;
	}

	virtual ssize_t send_next_chunk(int sock_fd);
};

#endif // CHUNKEDDATASENDER_H
