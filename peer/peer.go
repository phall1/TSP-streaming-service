package main

import (
	"encoding/gob"
	"fmt"
	"github.com/tcnksm/go-input"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	// "path/filepath"
	"strconv"
	"strings"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
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
	PAUSE
)

//const TRACKER_IP = "172.17.31.37:"
// const TRACKER_IP = "10.41.6.198:"

const TRACKER_IP = "172.17.92.155:"

var master_list string
var local_list = make([]string, 0)

// tracker to update master list, we then
// have to send updated list to all peers
// without have to request it so all songs
// will be available? or will user always
// lsit before playing? cause list updates
// songs think so this wouldnt be an
// issue.

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
	go serve_songs(args)

	for {
		if handle_command(args) < 0 {
			break
		}
	}
}

/**
* prints master list received from tracker
 */
func print_master_list(list string) {
	// TODO: format output nicely
	rows := strings.Split(list, "\n")
	for _, r := range rows {
		r = strings.Replace(r, ", ", "\t", -1)
		end := strings.Index(r, ">")
		fmt.Println(r[:end])
	}
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	// Check the address type and if it is not a loopback then display it
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

/**
* send local songs to server,
* tracker other peers now have this ip
 */
func become_discoverable(args []string) {
	songs := get_local_song_info(args[2])

	// Connect to tacker
	fmt.Println(TRACKER_IP + args[1])
	tracker, err := net.Dial("tcp", TRACKER_IP+args[1])
	if err != nil {
		panic(err)
		// fmt.Println("Error connecting to tracker")
		// os.Exit(1)
	}
	defer tracker.Close()

	send_local_song_info(tracker, songs)
}

/**
*
 */
func serve_songs(args []string) {
	server_ln, err := net.Listen("tcp", GetLocalIP()+":"+args[1])
	if err != nil {
		panic(err)
	}
	defer server_ln.Close()

	for {
		client, err := server_ln.Accept()
		if err != nil {
			continue
		}
		go receive_message(client)
	}
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

		// var extension = filepath.Ext(info_files[i].Name())
		// var mp3_name = "> "
		// mp3_name += info_files[i].Name()[0 : len(info_files[i].Name())-len(extension)]
		//
		// content = append(content, mp3_name...)
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
	cmd := get_cmd()
	hdr := new(TSP_header)
	var peer_ip string

	switch cmd {
	case "LIST":
		hdr.Type = LIST
		fmt.Println("LIST")
		tracker := send_request(*hdr, TRACKER_IP+args[1])
		receive_master_list(tracker)
	case "PLAY":
		hdr.Type = PLAY
		fmt.Println("PLAY")
		hdr.Song_id, peer_ip = get_song_selection()
		peer := send_play_request(*hdr, peer_ip+args[1], 69)
		go receive_mp3(peer)
		// receive mp3
		// play mp3
	case "PAUSE":
		hdr.Type = PAUSE
		fmt.Println("PAUSE")
		// find song playing and stop if (in the other goroutine, use
		// channesl)
	case "QUIT":
		hdr.Type = QUIT
		_ = send_request(*hdr, TRACKER_IP+args[1])
		fmt.Println("QUIT")
		// close all connections and quit
		return -1
	default:
		fmt.Println("invalid command")
	}
	return 0
}

func get_cmd() string {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
	query := "Select option"
	cmd, _ := ui.Select(query, []string{"LIST", "PLAY", "PAUSE", "QUIT"}, &input.Options{
		Loop: true,
	})
	return cmd
}

func get_song_selection() (int, string) {
	songs := strings.Split(master_list, "\n")
	var ip string

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
	query := "Select a song"
	id, _ := ui.Ask(query, &input.Options{
		ValidateFunc: func(id string) error {
			for _, s := range songs {
				if strings.Contains(s, id) {
					ip = strings.SplitN(s, ":", 3)[1][1:]
					return nil
				}
			}
			return fmt.Errorf("song id not here")
		},
		Loop: true,
	})
	ret, _ := strconv.ParseInt(id, 10, 32)
	return int(ret), ip + ":"
}

func send_play_request(hdr TSP_header, ip string, song_id int) net.Conn {

	peer, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Println("Error connecting to peer")
		return nil
	}

	tmp_msg := "DUMMY MESSAGE: HELLO"
	encoder := gob.NewEncoder(peer)
	msg_struct := &TSP_msg{hdr, []byte(tmp_msg)}
	encoder.Encode(msg_struct)
	return peer
}

func send_request(hdr TSP_header, dest_ip string) (conn net.Conn) {
	conn, err := net.Dial("tcp", dest_ip)
	if err != nil {
		fmt.Println("error connecting to " + dest_ip)
		os.Exit(1)
	}
	msg_content := ""
	encoder := gob.NewEncoder(conn)
	msg_struct := &TSP_msg{hdr, []byte(msg_content)}
	encoder.Encode(msg_struct)
	return
}

/**
* receives master list from tracker
* prints master list received from tracker
 */
func receive_master_list(tracker net.Conn) {
	defer tracker.Close()
	decoder := gob.NewDecoder(tracker)
	in_msg := new(TSP_msg)
	decoder.Decode(&in_msg)

	master_list = string(in_msg.Msg[:])
	print_master_list(master_list)
}

func get_song_filename(id string) string {
	rows := strings.Split(master_list, "\n")
	for _, r := range rows {
		if strings.Split(r, ":")[0] == id {
			r = strings.Replace(r, ", ", "\t", -1)
			end := strings.Index(r, ">")
			return r[end+2:]
		}
	}
	return ""
}

/**
 * called by the server thread, will act accordingly
 */
func receive_message(client net.Conn) {
	decoder := gob.NewDecoder(client)
	in_msg := new(TSP_msg)
	decoder.Decode(&in_msg)

	switch in_msg.Header.Type {
	case PLAY:
		song_file := get_song_filename(strconv.Itoa(in_msg.Header.Song_id))
		send_mp3_file(song_file, client)
		if song_file == "" {
			fmt.Println("error sing file empty") // to silence the warnings
		}
	default:
		// send a null response
		return
	}
}

func send_mp3_file(song_file string, client net.Conn) {
	defer client.Close()

	f, err := os.Open("songs/" + song_file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = io.Copy(client, f)
	if err != nil {
		panic(err)
	}
}

func receive_mp3(server net.Conn) {
	decoder, err := mp3.NewDecoder(server)
	if err != nil {
		panic(err)
	}
	defer decoder.Close()
	player, err := oto.NewPlayer(decoder.SampleRate(), 2, 2, 8192)
	if err != nil {
		panic(err)
	}
	defer player.Close()

	fmt.Printf("Length: %d[bytes]\n", decoder.Length())

	if _, err := io.Copy(player, decoder); err != nil {
		panic(err)
	}
}
