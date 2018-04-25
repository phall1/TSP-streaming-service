package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	// "github.com/gorilla/websocket"
)

const (
	INFO_FILE = "songs.info"
)

func main() {
	args := os.Args[:]
	if len(args) != 2 {
		fmt.Println("Usage: ", args[0], "<port>")
		os.Exit(1)
	}

	// setup server socket
	ln, err := net.Listen("tcp", ":"+args[1])
	if err != nil {
		panic(err)
	}
	// defer ln.Close() //close connection when function is done

	for {
		conn, err := ln.Accept()
		if err != nil {
			// error accepting this connection
			continue
		}
		go handleConnection(conn)
	}
}

func write_songs_to_info(peer net.Conn, song_bytes []byte) {
	// lock
	song_strs := strings.Split(string(song_bytes[:]), "\n")
	ip := peer.RemoteAddr().String() + ", "

	// info_file, err := os.OpenFile(INFO_FILE, os.O_APPEND, 0666)
	info_file, err := os.OpenFile(INFO_FILE, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		panic("cant open songs.info file")
	}
	defer info_file.Close()

	for _, s := range song_strs {
		record := "ID, " + ip + s
		info_file.WriteString(record + "\n")
		info_file.Sync()
	}
	// unlock
}

func handleConnection(peer net.Conn) {

	// info_bytes, err:= ioutil.ReadAll(peer)
	buff := make([]byte, 4096)
	bytes_read, err := peer.Read(buff)
	if err != nil {
		fmt.Println("error reading song")
		os.Exit(1)
	}

	buff = append(buff[:bytes_read])

	write_songs_to_info(peer, buff)
	os.Exit(1)

	// get songs from client
	// send out info about songs to all hosts

	// receive songs from client

	// update info doc (locks)

	// send client (all?) updated doc
}
