package main

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

type TSP_header struct {
	Type    byte
	Song_id int
}

type TSP_msg struct {
	Header TSP_header
	Msg    []byte
}

const (
	INIT = iota
	LIST
)

const INFO_FILE = "songs.info"

func init() {
	gob.Register(&TSP_header{})
	gob.Register(&TSP_msg{})
}

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
	defer ln.Close()

	for {
		peer, err := ln.Accept()
		if err != nil {
			// error accepting this connection
			fmt.Println("error accepting conn")
			continue
		}
		go handleConnection(peer)
	}
}

/**
 * takes new peer's song list, and adds their songs
 * to the info file, with the peers IP addess, and
 * assigns ID's to the new songs
 */
func write_songs_to_info(peer net.Conn, song_bytes []byte) {
	// lock
	song_strs := strings.Split(string(song_bytes[:]), "\n")
	ip := peer.RemoteAddr().String() + ", "

	info_file, err := os.OpenFile(INFO_FILE, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		panic("cant open songs.info file")
	}
	defer info_file.Close()

	for _, s := range song_strs {
		if s == "" {
			continue
		}
		record := "ID, " + ip + s
		info_file.WriteString(record + "\n")
		info_file.Sync()
	}
	// unlock
}

/**
 * handles new connections
 * either takes the new peer's song list,
 * or sends back the master info file
 */
func handleConnection(peer net.Conn) {
	decoder := gob.NewDecoder(peer)
	in_msg := new(TSP_msg)
	decoder.Decode(&in_msg)

	switch in_msg.Header.Type {
	case INIT:
		fmt.Println("INIT")
		write_songs_to_info(peer, in_msg.Msg)
	case LIST:
		fmt.Println("INFO")
		send_info_file(peer)
	default:
		fmt.Println("Bad Msg Header")
	}
	peer.Close()
}

/**
 * send the master song info file to the peer
 * that requested it
 */
func send_info_file(peer net.Conn) {
	info_file, err := ioutil.ReadFile(INFO_FILE)
	if err != nil {
		panic("cant open songs.info file")
	}

	encoder := gob.NewEncoder(peer)
	msg_struct := &TSP_msg{TSP_header{Type: 1}, []byte(info_file)}
	encoder.Encode(msg_struct)
}
