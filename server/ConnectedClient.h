#ifndef CONNECTEDCLIENT_H
#define CONNECTEDCLIENT_H

enum ClientState { RECEIVING, SENDING };

class ConnectedClient {
  public:
	int client_fd;
	ChunkedDataSender *sender;
	ClientState state;

	/**
	 * Sends a response to the client.
	 * Note that this is just to demonstrate sending to the client: it doesn't
	 * send anything intelligent.
	 *
	 * @param epoll_fd File descriptor for epoll.
	 */
	void send_dummy_response(int epoll_fd);

	/**
	 * Handles new input from the client.
	 *
	 * @param epoll_fd File descriptor for epoll.
	 */
	void handle_input(int epoll_fd);

	/**
	 * Handles a close request from the client.
	 *
	 * @param epoll_fd File descriptor for epoll.
	 */
	void handle_close(int epoll_fd);
};

#endif
