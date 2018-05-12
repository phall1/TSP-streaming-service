package main

import (
	"encoding/gob"
	"fmt"
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
	INFO
	PLAY
	STOP
	QUIT

	MAX_SONGS = 1000
)

var (
	id_counter int = 10
	info           = make([]string, 0)
)

/*
 * Allows data to be sent using Gob
 */
func init() {
	gob.Register(&TSP_header{})
	gob.Register(&TSP_msg{})
}

/*
 * Return the tracker server's IP address.
 *
 * @return the IP address in string form
 */

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

	// Setup server socket
	ln, err := net.Listen("tcp", GetLocalIP()+":"+args[1])
	// ln, err := net.Listen("tcp", "localhost:"+args[1])
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	var mutex = &sync.Mutex{}
	for {
		peer, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting conn")
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
 *
 * @param peer Connection with a peer on network
 * @param mutex Mutex for locking master song list
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
 * @param peer Peer connectoin
 * @param song_bytes the bytes containing song info
 */
func get_info_from_peer(peer net.Conn, song_bytes []byte) {
	song_strs := strings.Split(string(song_bytes[:]), "\n")
	ip := peer.RemoteAddr().String() + ", "
	for i, _ := range song_strs {
		if song_strs[i] == "" {
			continue
		}
		info = append(info, strconv.Itoa(id_counter)+": "+ip+song_strs[i])
		id_counter++
	}
	fmt.Println(info)
}

/**
 * takes new peer's song list, and adds their songs
 * to the info file, with the peers IP addess, and
 * assigns ID's to the new songs
 * @param peer the Peer connection
 */
func remove_songs(peer net.Conn) {
	ip := peer.RemoteAddr().String()
	ip_slice := strings.Split(ip, ":")
	for i := 0; i < len(info); i++ {
		if strings.Contains(info[i], ip_slice[0]) {
			info = append(info[:i], info[i+1:]...)
			i--
		}
	}
	fmt.Println(info)
}

/**
 * send the master song info file to the peer
 * that requested it
 * @param peer the Peer connection
 */
func send_info_file(peer net.Conn) {
	info_msg := strings.Join(info, "\n")
	encoder := gob.NewEncoder(peer)
	encoder.Encode(&TSP_msg{TSP_header{Type: 1}, []byte(info_msg)})
}
