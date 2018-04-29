package main

import (
	"encoding/gob"
	"fmt"
	// "io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
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
	PLAY
	QUIT
)

const MAX_SONGS = 1000

var info = make([]string, 0)
var id_counter int = 10

func init() {
	gob.Register(&TSP_header{})
	gob.Register(&TSP_msg{})
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func main() {
	args := os.Args[:]
	if len(args) != 2 {
		fmt.Println("Usage: ", args[0], "<port>")
		os.Exit(1)
	}
	fmt.Println(GetLocalIP())

	// setup server socket
	ln, err := net.Listen("tcp", GetLocalIP()+":"+args[1])
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	var mutex = &sync.Mutex{}

	for {
		peer, err := ln.Accept()
		if err != nil {
			// error accepting this connection
			fmt.Println("error accepting conn")
			continue
		}
		fmt.Println("handle_connection")
		go handleConnection(peer, mutex)
	}
}

/**
 * handles new connections
 * either takes the new peer's song list,
 * or sends back the master info file
 */
func handleConnection(peer net.Conn, mutex *sync.Mutex) {
	defer peer.Close()
	decoder := gob.NewDecoder(peer)
	in_msg := new(TSP_msg)
	decoder.Decode(&in_msg)

	mutex.Lock()
	switch in_msg.Header.Type {
	case INIT:
		fmt.Println("INIT")
		get_info_from_peer(peer, in_msg.Msg)
	case LIST:
		fmt.Println("INFO")
		send_info_file(peer)
	case QUIT:
		fmt.Println("QUIT")
		// remove songs in the lsit that was jsut sent
		// Mimic the INIT case?
		remove_songs(peer)
	default:
		fmt.Println("Bad Msg Header")
	}
	mutex.Unlock()
}

/**
 * takes new peer's song list, and adds their songs
 * to the info file, with the peers IP addess, and
 * assigns ID's to the new songs
 */
func get_info_from_peer(peer net.Conn, song_bytes []byte) {
	// TODO: lock
	// TODO: assign id numbers properly
	song_strs := strings.Split(string(song_bytes[:]), "\n")
	ip := peer.RemoteAddr().String() + ", "
	fmt.Println("loop")
	for i := 0; i < len(song_strs)-1; i += 2 {
		// for i, s := range song_strs {
		if song_strs[i] == "" {
			continue
		}
		fmt.Println(id_counter)
		record := strconv.Itoa(id_counter)
		record += ": "
		record += ip
		record += song_strs[i]
		record += song_strs[i+1]
		i++

		// fmt.Println(s)
		// record := "ID: " + ip + s
		id_counter++
		info = append(info, record)
	}
	fmt.Println("end loop")
	fmt.Println(info)
	// unlock
}

/**
 * takes new peer's song list, and adds their songs
 * to the info file, with the peers IP addess, and
 * assigns ID's to the new songs
 */
func remove_songs(peer net.Conn) {
	// TODO: lock
	// TODO: assign id numbers properly
	ip := peer.RemoteAddr().String()
	ip_slice := strings.Split(ip, ":")

	for i := 0; i < len(info); i++ {
		if strings.Contains(info[i], ip_slice[0]) {
			info = append(info[:i], info[i+1:]...)
			i--
		}
	}

	fmt.Println(info)
	// unlock
}

/**
 * send the master song info file to the peer
 * that requested it
 */
func send_info_file(peer net.Conn) {
	info_msg := strings.Join(info, "\n")
	encoder := gob.NewEncoder(peer)
	msg_struct := &TSP_msg{TSP_header{Type: 1}, []byte(info_msg)}
	encoder.Encode(msg_struct)
}
