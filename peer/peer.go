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
	STOP
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
	}
	defer tracker.Close()
	send_local_song_info(tracker, songs)
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

func get_cmd() string {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
	query := "Select option"
	cmd, _ := ui.Select(query, []string{"LIST", "PLAY", "STOP", "QUIT"}, &input.Options{
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

func main() {
	args := os.Args[:]
	if len(args) != 3 {
		fmt.Println("Usage: ", args[0], "<port> <filedir>")
		os.Exit(1)
	}

	become_discoverable(args)

	// go start serving songs
	//TODO: EPOLL
	go serve_songs(args)

	send_stop_chan := make(chan struct{})
	play := make(chan bool)
	stop := make(chan bool)

	for {
		if handle_command(args, send_stop_chan, play, stop) < 0 {
			break
		}
	}
}

/**
* handle input command from the user
* LIST - get song list from peers
* PLAY <song id> - play song
* PAUSE - pauses playing of song (buffering continues)
* STOP - stop streaming song
* QUIT - <--
 */
func handle_command(args []string, send_stop_chan chan struct{}, play chan bool, stop chan bool) int {
	cmd := get_cmd()
	hdr := new(TSP_header)
	var peer_ip string
	//var peer net.Conn

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
		go receive_mp3(peer, send_stop_chan, play, stop)
		play <- true
	case "STOP":
		hdr.Type = STOP
		fmt.Println("STOP")
		stop <- true
		//peer.Close()
		// close(send_stop_chan)
		// <-send_stop_chan
		// fmt.Println("Closed stop chan waiting for done chan")
		// fmt.Println("done chan")
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

func receive_mp3(server net.Conn, send_stop_chan chan struct{}, play chan bool, stop chan bool) {
	fmt.Println("here")
	for {
		select {
		// case _ = <-send_stop_chan:
		case <-stop:
			// fmt.Println("closgin player")
			return
		case <-play:
			// default:
			decoder, err := mp3.NewDecoder(server)
			if err != nil && err == io.EOF {
				// fmt.Println("EOF error handled.")
				return
			}
			if err != nil && err != io.EOF {
				panic(err)
			}
			defer decoder.Close()
			player, err := oto.NewPlayer(decoder.SampleRate(), 2, 2, 8192)
			if err != nil && err == io.EOF {
				// fmt.Println("EOF error handled.")
				return
			}
			if err != nil && err != io.EOF {
				panic(err)
			}
			defer player.Close()

			go func() {
				if _, _ = io.Copy(player, decoder); err != nil {
					// fmt.Println(err)
					// panic(err)
				}
			}()
		}
	}
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

func serve_songs_epoll(args []string) {
	server_ln, err := net.Listen("tcp", GetLocalIP()+":"+args[1])
	//var event syscall.EpollEvent
	//var events[MAX_EPOLL_EVENTS]syscall.EpollEvent

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

	_, _ = io.Copy(client, f)
	// if err != nil {
	// fmt.Println(err)
	// panic(err)
	// }
}

func play_mp3(server net.Conn) {
	decoder, err := mp3.NewDecoder(server)
	if err != nil && err == io.EOF {
		fmt.Println("EOF error handled.")
		return
	}
	if err != nil && err != io.EOF {
		panic(err)
	}
	defer decoder.Close()
	player, err := oto.NewPlayer(decoder.SampleRate(), 2, 2, 8192)
	if err != nil && err == io.EOF {
		fmt.Println("EOF error handled.")
		return
	}
	if err != nil && err != io.EOF {
		panic(err)
	}
	defer player.Close()

	if _, err := io.Copy(player, decoder); err != nil {
		panic(err)
	}
}
