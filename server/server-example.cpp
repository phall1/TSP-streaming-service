/*
 * File: server-example.c
 *
 * Authors: (TODO: Fill this in with author info)
 *
 * Example code to help get you starter with your Jukebox server.
 * 
 * DO NOT MODIFY THIS FILE!!!
 *
 * Instead, make a copy of it named jukebox-server.c and modify that file.
 */

// C++ standard libraries
#include <string>
#include <iostream>
#include <fstream>
#include <sstream>

// C++ extra libraries
#include <boost/filesystem.hpp>

// C standard libraries
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <cerrno>

// POSIX and OS-specific libraries
#include <unistd.h>
#include <dirent.h>
#include <fcntl.h>
#include <inttypes.h>
#include <arpa/inet.h>
#include <netinet/in.h>
#include <sys/epoll.h>
#include <sys/socket.h>
#include <sys/time.h>
#include <sys/types.h>

#define BACKLOG (10)
#define MAX_CLIENTS (1024)
#define MAX_EVENTS (1024)

namespace fs = boost::filesystem;

// forward declarations
int acceptConnection(int server_socket);
int setup_server_socket(uint16_t port_num);
void setNonBlocking(int sock);
int readMP3Files(char *dir);

/*
 * The main function for the server.
 * You should refactor this to be smaller and add additional functionality as
 * needed.
 */
int main(int argc, char **argv) {
    if (argc != 3) {
        printf("Usage:\n%s <port> <filedir>\n", argv[0]);
        exit(0);
    }

	if (!fs::is_directory(argv[2])) {
		printf("ERROR: %s is not a directory\n", argv[2]);
		exit(1);
	}

    // Get the port number from the arguments.
    uint16_t port = (uint16_t) std::stoul(argv[1]);

	int serv_sock = setup_server_socket(port);

    /* 
	 * Read the other argument (mp3 directory).
	 * See the notes for this function above.
	 */
    int song_count = readMP3Files(argv[2]);
    printf("Found %d songs.\n", song_count);

	// Create the epoll, which returns a file descriptor for us to use later.
	int epoll_fd = epoll_create1(0);
	if (epoll_fd < 0) {
		perror("epoll_create1");
		exit(1);
	}

	// We want to watch for input events (i.e. connection requests) on our
	// server socket.
	struct epoll_event server_ev;
	server_ev.data.fd = serv_sock;
	server_ev.events = EPOLLIN;

	if (epoll_ctl(epoll_fd, EPOLL_CTL_ADD, serv_sock, &server_ev) == -1) {
		perror("epoll_ctl");
		exit(1);
	}


    while (true) {
		// wait for some events to occur, writing them to our events array
		struct epoll_event events[MAX_EVENTS];

		int num_events = epoll_wait(epoll_fd, events, MAX_EVENTS, -1);
		if (num_events < 0) {
			perror("epoll_wait");
			exit(1);
		}

		// Loop through all the I/O events that just happened.
		for (int n = 0; n < num_events; n++) {
			// Check if this is a "hang up" event
			if ((events[n].events & EPOLLRDHUP) != 0) {
				// If we get here, the socket associated with this event was
				// closed by the remote host so we should just call close here
				// and remove this event using epoll_ctl with EPOLL_CTL_DEL.
			}

			// Check if this is an "input" event
			if ((events[n].events & EPOLLIN) != 0) {
				if (events[n].data.fd == serv_sock) {

					// If the server socket is ready for "reading," that implies
					// we have a new client that wants to connect so lets
					// accept it.
					int client_fd = acceptConnection(serv_sock);
					printf("Accepted a new connection!\n");
					
					// Watch for input and output events and "hangup" events
					// for new clients.
					struct epoll_event new_client_ev;
					new_client_ev.events = EPOLLIN | EPOLLOUT | EPOLLRDHUP;
					new_client_ev.data.fd = client_fd;

					if (epoll_ctl(epoll_fd, EPOLL_CTL_ADD, client_fd, 
									&new_client_ev) == -1) {
						perror("epoll_ctl: client_fd");
						exit(1);
					}

					// Set this to non-blocking mode so we never get hung up
					// trying to send or receive from this client.
					setNonBlocking(client_fd);
				}
				else {
					// This wasn't the server socket so this means we have a
					// client that has sent us data so we can receive it now
					// without worrying about blocking.
				}
            }

			// Check if this is an "output" event
			if ((events[n].events & EPOLLOUT) != 0) {
				// The socket associated wih this event is ready for writing
				// (i.e. ready for us to send data to it).

                 /* 
				  * I would recommend calling another function here rather
                  * than cluttering up main and this select loop.
				  */
            }
        }
    }
}

/**
 * Creates a socket, sets it to non-blocking, binds it to the given port, then
 * sets it to start listen for incoming connections.
 *
 * @param port_num The port number we will listen on.
 * @return The file descriptor of the newly created/setup server socket.
 */
int setup_server_socket(uint16_t port_num) {
    /* Create the socket that we'll listen on. */
    int sock_fd = socket(AF_INET, SOCK_STREAM, 0);

    /* Set SO_REUSEADDR so that we don't waste time in TIME_WAIT. */
    int val = setsockopt(sock_fd, SOL_SOCKET, SO_REUSEADDR, 
							&val, sizeof(val));
    if (val < 0) {
        perror("Setting socket option failed");
        exit(1);
    }

    /* 
	 * Set our server socket to non-blocking mode.  This way, if we
     * accidentally accept() when we shouldn't have, we won't block
     * indefinitely.
	 */
    setNonBlocking(sock_fd);

    struct sockaddr_in addr;
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port_num);
    addr.sin_addr.s_addr = INADDR_ANY;

    /* Bind our socket and start listening for connections. */
    if (bind(sock_fd, (struct sockaddr*)&addr, sizeof(addr)) < 0) {
        perror("Error binding to port");
        exit(1);
    }

    if (listen(sock_fd, BACKLOG) < 0) {
        perror("Error listening for connections");
        exit(1);
    }

	return sock_fd;
}

/**
 * Accepts a connection and returns the socket descriptor of the new client
 * that has connected to us.
 *
 * @param server_socket Socket descriptor of the server (that is listening)
 * @return Socket descriptor for newly connected client.
 */
int acceptConnection(int server_socket) {
	struct sockaddr_storage their_addr;
	socklen_t addr_size = sizeof(their_addr);
	int new_fd = accept(server_socket, (struct sockaddr *)&their_addr,
			&addr_size);
	if (new_fd < 0) {
		perror("accept");
		exit(1);
	}

	return new_fd;
}

/* 
 * Use fcntl (file control) to set the given socket to non-blocking mode.
 *
 * @info Setting your sockets to non-blocking mode is not required, but it
 * might help with your debugging.  By setting each socket you get from
 * accept() to non-blocking, you can be sure that normally blocking calls like
 * send, recv, and accept will instead return an error condition and set errno
 * to EWOULDBLOCK/EAGAIN.  I would recommend that you set your sockets for
 * non-blocking and then explicitly check each call to send, recv, and accept
 * for this errno.  If you see it happening, you know that you're attempting
 * to call one of those functions when you shouldn't be.
 *
 * @param sock The file descriptor for the socket you want to make
 * 				non-blocking.
 */
void setNonBlocking(int sock) {
    /* Get the current flags. We want to add O_NONBLOCK to this set. */
    int socket_flags = fcntl(sock, F_GETFD);
    if (socket_flags < 0) {
        perror("fcntl");
        exit(1);
    }

    /* Add in the O_NONBLOCK flag by bitwise ORing it to the old flags. */
    socket_flags = socket_flags | O_NONBLOCK;

    /* Set the new flags, including O_NONBLOCK. */
    int result = fcntl(sock, F_SETFD, socket_flags);
    if (result < 0) {
        perror("fcntl");
        exit(1);
    }

    /* The socket is now in non-blocking mode. */
}


/*
 * Given a path to a directory, this function searches the directory for any
 * files that end in ".mp3".
 * When it files an MP3 file, it also looks for an associated "info" file, and
 * prints the contents of this file if it exists.
 *
 * @info You'll need to edit this to meet your needs (i.e. don't expect this
 * to do everything you want without any effort).
 *
 * @param dir String that represents the path to the directory that you want
 * 				to check.
 *
 * @return Number of MP3 files found inside of the specified directory.
 */
int readMP3Files(char *dir) {
	int num_mp3_files = 0;

	// Loop through all files in the directory
	for(fs::directory_iterator entry(dir); entry != fs::directory_iterator(); ++entry) {
		std::string filename = entry->path().filename().string();

		// See if the current file is an MP3 file
		if (entry->path().extension() == ".mp3") {
			printf("(%d) %s\n", num_mp3_files, filename.c_str());
			num_mp3_files++;

			// Look for an associated info file
			fs::path info_file_path = entry->path();
			info_file_path = info_file_path.replace_extension(".mp3.info");
			if (fs::is_regular_file(info_file_path)) {
				// read contents of file into a buffer
				std::ifstream t(info_file_path.string());
				std::stringstream buffer;
				buffer << t.rdbuf();
				printf("Info:\n%s\n\n", buffer.str().c_str());
			}
		}
	}

    return num_mp3_files;
}
