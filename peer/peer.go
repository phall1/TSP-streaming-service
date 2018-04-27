package main

import (
	"encoding/gob"
	"fmt"
	"github.com/tcnksm/go-input"
	"io/ioutil"
	"net"
	"os"
	"path"
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
	PLAY
	PAUSE
	QUIT
)

func init() {
	gob.Register(&TSP_header{})
	gob.Register(&TSP_msg{})
}

func main() {
	args := os.Args[:]
	if len(args) != 3 {
		fmt.Println("Usage: ", args[0], "<port> <filedir>")
		os.Exit(1)
	}

	become_discoverable(args)

	// go start serving songs

	for {
		if handle_command(args) < 0 {
			break
		}
	}
}

/**
* send local songs to server,
* tracker other peers now have this ip
 */
func become_discoverable(args []string) {
	songs := get_local_song_info(args[2])

	// Connect to tacker
	tracker, err := net.Dial("tcp", "localhost:"+args[1])
	if err != nil {
		fmt.Println("Error connecting to tracker")
		os.Exit(1)
	}
	defer tracker.Close()

	send_local_song_info(tracker, songs)

	// tracker_port int = args[2]
}

/**
* Returns an string slice
* for all files(songs) in directory "songs"
 */
func get_local_song_info(dir_name string) []string {
	info_files, err := ioutil.ReadDir(dir_name)
	if err != nil {
		fmt.Println("cant read songs")
		os.Exit(1)
	}

	song_info := make([]string, len(info_files))
	for i := 0; i < len(info_files); i++ {
		if path.Ext(info_files[i].Name()) != ".info" {
			continue
		}
		content, _ := ioutil.ReadFile(dir_name + "/" + info_files[i].Name())
		song_info = append(song_info, string(content[:]))
	}
	return song_info
}

/**
* send this peers song list to the
* tracker server using gobs to transfer structs
 */
func send_local_song_info(tracker net.Conn, songs []string) {
	msg_content := ""
	for _, s := range songs {
		msg_content += s
	}

	encoder := gob.NewEncoder(tracker)
	msg_struct := &TSP_msg{TSP_header{Type: 0, Song_id: 0}, []byte(msg_content)}
	encoder.Encode(msg_struct)
}

/**
* handle input command from the user
* LIST - get song list from peers
* PLAY <song id> - play song
* PAUSE - pauses playing of song (buffering continues)
* STOP - stop streaming song
* QUIT - <--
 */
func handle_command(args []string) int {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
	query := "Select option"
	cmd, _ := ui.Select(query, []string{"LIST", "PLAY", "PAUSE", "QUIT"}, &input.Options{
		Loop: true,
	})

	hdr := new(TSP_header)

	switch cmd {
	case "LIST":
		// get list from server
		hdr.Type = LIST
		fmt.Println("LIST")
		tracker := send_list_request(*hdr, args) // TSP_header error
		receive_master_list(tracker)
	case "PLAY":
		hdr.Type = PLAY
		fmt.Println("PLAY")
		// go play song
		play_song(*hdr, "localhost:6969", 69)
	case "PAUSE":
		hdr.Type = PAUSE
		fmt.Println("PAUSE")
		// find song playing and stop if (in the other goroutine, use
		// channesl)
	case "QUIT":
		hdr.Type = QUIT
		fmt.Println("QUIT")
		// close all connections and quit
		return -1
	default:
		fmt.Println("invalid command")
	}
	//fmt.Println("after switch")
	return 0
}

func ask_for_id(list []string) {

}

func play_song(hdr TSP_header, ip string, song_id int) {
	peer, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Println("Error connecting to peer")
		return
	}

	tmp_msg := "play song here"
	encoder := gob.NewEncoder(peer)
	msg_struct := &TSP_msg{hdr, []byte(tmp_msg)}
	encoder.Encode(msg_struct)
}

/**
* sends a request for list of songs
* available on the network
 */
func send_list_request(hdr TSP_header, args []string) net.Conn { // parameter error
	tracker, err := net.Dial("tcp", "localhost:"+args[1])
	if err != nil {
		fmt.Println("Error connecting to tracker")
		os.Exit(1)
	}

	msg_content := ""
	encoder := gob.NewEncoder(tracker)
	msg_struct := &TSP_msg{TSP_header{Type: 1}, []byte(msg_content)}
	encoder.Encode(msg_struct)
	return tracker
}

/**
* receives master list from tracker
 */
func receive_master_list(tracker net.Conn) {
	defer tracker.Close()
	decoder := gob.NewDecoder(tracker)
	in_msg := new(TSP_msg)
	decoder.Decode(&in_msg)

	print_master_list(string(in_msg.Msg[:]))
}

/**
* prints master list received from tracker
 */
func print_master_list(list string) {
	// TODO: format output nicely
	rows := strings.Split(list, "\n")
	for _, r := range rows {
		r = strings.Replace(r, ", ", "\t", -1)
		// row_slice := strings.Split(r, ",")
		// for _, e := range row_slice {
		//     fmt.Printf("%")
		// }
		//
		fmt.Println(r)
	}
}
