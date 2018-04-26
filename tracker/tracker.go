package main

import (
	// "bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"
	// "github.com/gorilla/websocket"
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
	INFO_FILE = "songs.info"
	INIT      = 0
	LIST      = 1
)

func init() {
	// This type must match exactly what youre going to be using,
	// down to whether or not its a pointer
	// FUCK THIS
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

	decoder := gob.NewDecoder(peer)
	in_msg := new(TSP_msg)
	decoder.Decode(&in_msg)
	fmt.Println(in_msg)

	switch in_msg.Header.Type {
	case 0:
		write_songs_to_info(peer, in_msg.Msg)
	case 1:
		fmt.Println("case 1 ")
	default:
		fmt.Println("didnt work")
	}
	// in_hdr := in_msg.Header
	peer.Close()
	os.Exit(1)
}
