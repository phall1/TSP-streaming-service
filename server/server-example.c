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

#include <arpa/inet.h>
#include <dirent.h>
#include <errno.h>
#include <fcntl.h>
#include <inttypes.h>
#include <netinet/in.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/select.h>
#include <sys/socket.h>
#include <sys/time.h>
#include <sys/types.h>
#include <unistd.h>

#define BACKLOG (10)
#define MAX_CLIENTS (1024)

// forward declarations
int acceptConnection(int server_socket);
void setNonBlocking(int sock);
int filter(const struct dirent *ent);
int readMP3Files(char *dir);

/*
 * The main function for the server.
 * You should refactor this to be smaller and add additional functionality as
 * needed.
 */
int main(int argc, char **argv) {
    if (argc < 3) {
        printf("Usage:\n%s <port> <filedir>\n", argv[0]);
        exit(0);
    }

    // Get the port number from the arguments.
    uint16_t port = (uint16_t) atoi(argv[1]);

    // Create the socket that we'll listen on.
    int serv_sock = socket(AF_INET, SOCK_STREAM, 0);

    /* 
	 * Set SO_REUSEADDR so that we don't waste time in TCP's TIME_WAIT when we
	 * shut down the connection.
	 */
    int val = setsockopt(serv_sock, SOL_SOCKET, SO_REUSEADDR, &val,
            sizeof(val));

    if (val < 0) {
        perror("setsockopt");
        exit(1);
    }

    /* Set our server socket to non-blocking mode.  This way, if we
     * accidentally accept() when we shouldn't have, we won't block
     * indefinitely. */
    setNonBlocking(serv_sock);

    /* 
	 * Read the other argument (mp3 directory).
	 * See the notes for this function above.
	 */
    int song_count = readMP3Files(argv[2]);
    printf("Found %d songs.\n", song_count);

    struct sockaddr_in addr;
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    addr.sin_addr.s_addr = INADDR_ANY;

    // Bind our socket and start listening for connections.
    val = bind(serv_sock, (struct sockaddr*)&addr, sizeof(addr));
    if(val < 0) {
        perror("bind");
        exit(1);
    }

	// Set our socket to listen for connections (which we'll later accept)
    val = listen(serv_sock, BACKLOG);
    if(val < 0) {
        perror("listen");
        exit(1);
    }

	/*
	 * Select uses sets of file descriptors (type: fd_set) to determine which
	 * file descriptors it will monitor.
	 */
    fd_set all_fds, read_fds, write_fds;
	
	FD_ZERO(&all_fds);
	FD_ZERO(&read_fds);
	FD_ZERO(&write_fds);

	/* 
	 * Start by adding the server socket to the list of all socket file
	 * descriptors.
	 * We'll check that this is available for reading to see if we have a new
	 * client trying to connect, which will force us to call accept().
	 */
	FD_SET(serv_sock, &all_fds);

	/*
	 * Keep track of the maximum file descriptor (which is just an integer),
	 * which we'll need when calling select.
	 */
	int max_fd = serv_sock;

    while (1) {

		/*
		 * This will cause us to look for reads or writes on all of our
		 * sockets.
		 * You may want to change this if it doesn't make sense for how you
		 * are communicating with clients.
		 */
		read_fds = all_fds;
		write_fds = all_fds;

		/* 
		 * select() will block until one of the sockets we asked about is ready for
		 * recving/sending.
		 * IMPORTANT NOTE: Select only checks up to, but not including, the
		 * value in the first param. For that reason, we'll add 1 to max_fd to
		 * ensure that max_fd actually gets checked.
		 */
        val = select(max_fd+1, &read_fds, &write_fds, NULL, NULL);

        if (val < 1) {
            perror("select");
            continue;
        }

        /* Check the FD_SETs after select returns.  These will tell us
         * which sockets are ready for recving/sending. */
        for (int i = 0; i <= max_fd; ++i) {
            if (FD_ISSET(i, &read_fds)) {
				if (i == serv_sock) {
					// If the server socket is ready for "reading," that implies
					// we have a new client that wants to connect so lets
					// accept it.
					int client_fd = acceptConnection(serv_sock);
					
					FD_SET(client_fd, &all_fds);
					if (client_fd > max_fd) max_fd = client_fd;
					printf("Accepted a new connection!\n");

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
            if (FD_ISSET(i, &write_fds)) {
                // Socket descriptor i is ready for writing.

                 /* 
				  * I would recommend calling another function here rather
                  * than cluttering up main and this select loop.
				  */
            }
        }
    }
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
 * Checks wheter a filename ends in '.mp3'. 
 *
 * @info This is used below in readMP3Files. You probably don't need to use
 * it anywhere else.
 *
 * @param ent Directory entity that we are going to check.
 *
 * @return 0 if ent doesn't end in .mp3, 1 if it does.
 */
int filter(const struct dirent *ent) {
    int len = strlen(ent->d_name);

    return !strncasecmp(ent->d_name + len - 4, ".mp3", 4);
}

/*
 * Given a path to a directory, this function scans that directory (using the
 * handy scandir library function) to produce an alphabetized list of files
 * whose names end in ".mp3".  For each one, it then also checks for a
 * corresponding ".info" file and reads that in its entirety.
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
    struct dirent **namelist;
    int i,n;

    n = scandir(dir, &namelist, filter, alphasort);
    if (n < 0) {
        perror("scandir");
        exit(1);
    }

    for (i = 0; i < n; ++i) {
        int bytes_read = 0;
        int total_read = 0;
        char path[1024];

        FILE *infofile = NULL;

        char *infostring = NULL;

        /* namelist[i]->d_name now contains the name of an mp3 file. */
        /* FIXME: You probably want to use this name to find file data. */
        printf("(%d) %s\n", i, namelist[i]->d_name);

        /* Build a path to a possible input file. */
        strcpy(path, dir);
        strcat(path, "/");
        strcat(path, namelist[i]->d_name);
        strcat(path, ".info");

        infofile = fopen(path, "r");
        if (infofile == NULL) {
            /* It wasn't there (or failed to open for some other reason). */
            infostring = "No information available.";
        } else {
            /* We found and opened the info file. */
            int infosize = 1024;
            infostring = malloc(infosize);

            do {
                infosize *= 2;
                infostring = realloc(infostring, infosize);

                bytes_read = fread(infostring + total_read, 1, infosize - total_read, infofile);
                total_read += bytes_read;
            } while (bytes_read > 0);

            fclose(infofile);

            /* Zero-out the unused space at the end of the buffer. */
            memset(infostring + total_read, 0, infosize - total_read);
        }

        /* infostring now contains the info data for this song. */
        /* FIXME: Use these info strings when clients send info commands. */
        printf("Info:%s\n\n", infostring);

        free(namelist[i]);
    }
    free(namelist);

    /* Return the number of files we found. */
    return n;
}
