/**
 * Authors: Patrick Hall
 *			James Ponwith
 *          Max Gradwohl
 */

package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
	"github.com/tcnksm/go-input"
)

const (
	INIT = iota
	LIST
	INFO
	PLAY
	STOP
	QUIT

	TRACKER_IP = "172.17.92.155:"
	/* TRACKER_IP = "172.17.31.37:" */
	MAX_EVENTS = 64
	EPOLLET    = 1 << 31
)

type TSP_header struct {
	Type    byte
	Song_id int
}

type TSP_msg struct {
	Header TSP_header
	Msg    []byte
}

type Reader struct {
	read string
	done bool
}

func NewReader(toRead string) *Reader {
	return &Reader{toRead, false}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.done {
		return 0, io.EOF
	}
	for i, b := range []byte(r.read) {
		p[i] = b
	}
	r.done = true
	return len(r.read), nil
}

var master_list string

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
	go serve_songs_epoll(args)

	play := make(chan bool)
	stop := make(chan bool)

	for {
		if handle_command(args, play, stop) < 0 {
			break
		}
	}
}

/*----------------------------SERVER----------------------------*/
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
/*
func receive_message(client net.Conn) {
	decoder := gob.NewDecoder(client)
	in_msg := new(TSP_msg)
	decoder.Decode(&in_msg)

	switch in_msg.Header.Type {
	case PLAY:
		song_file := get_song_filename(strconv.Itoa(in_msg.Header.Song_id))
		send_mp3_file(song_file, client)
		if song_file == "" {
			// fmt.Println("error sing file empty") // to silence the warnings
		}
	default:
		// send a null response
		return
	}
}
*/

/**
 * called by the server thread, will act accordingly
 */
func receive_message_epoll(client_fd int) {
	bytes := make([]byte, 1024)
	_, _ = syscall.Read(client_fd, bytes)

	M := NewReader(string(bytes))

	decoder := gob.NewDecoder(M)
	in_msg := new(TSP_msg)
	decoder.Decode(&in_msg)

	switch in_msg.Header.Type {
	case PLAY:
		song_file := get_song_filename(strconv.Itoa(in_msg.Header.Song_id))
		send_mp3_file(song_file, client_fd)
		if song_file == "" {
			// fmt.Println("error sing file empty") // to silence the warnings
		}
	default:
		// send a null response
		return
	}
}

/**
 * Server side of peers. Handle incoming peers, establish connections
 * and adventually send them music.
 */
/*
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
*/

func serve_songs_epoll(args []string) {
	// var event syscall.EpollEvent
	var event syscall.EpollEvent

	var events [MAX_EVENTS]syscall.EpollEvent

	fd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(fd)

	if err = syscall.SetNonblock(fd, true); err != nil {
		panic(err)
	}

	port, _ := strconv.ParseInt(args[1], 10, 32)
	addr := syscall.SockaddrInet4{Port: int(port)}
	copy(addr.Addr[:], net.ParseIP(GetLocalIP()).To4())

	syscall.Bind(fd, &addr)
	syscall.Listen(fd, 10)

	epfd, e := syscall.EpollCreate1(0)
	if e != nil {
		panic(e)
	}
	defer syscall.Close(epfd)

	event.Events = syscall.EPOLLIN
	event.Fd = int32(fd)
	if e = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &event); e != nil {
		panic(e)
	}

	for {
		nevents, e := syscall.EpollWait(epfd, events[:], -1)
		if e != nil {
			fmt.Println("epoll_wait: ", e)
			break
		}

		for ev := 0; ev < nevents; ev++ {
			if int(events[ev].Fd) == fd {
				connFd, _, err := syscall.Accept(fd)
				if err != nil {
					fmt.Println("accept: ", err)
					continue
				}
				syscall.SetNonblock(fd, true)
				event.Events = syscall.EPOLLIN | EPOLLET
				event.Fd = int32(connFd)
				err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, connFd, &event)
				if err != nil {
					panic(err)
				}
			} else {
				go receive_message_epoll(int(events[ev].Fd))
			}
		}
	}
}

func send_mp3_file(song_file string, client int) {
	defer syscall.Close(client)
	//f, err := os.Open("songs/" + song_file)
	bytes, err := ioutil.ReadFile("songs/" + song_file)
	if err != nil {
		panic(err)
	}
	//	defer f.Close()
	syscall.Write(client, bytes)
	//_, _ = io.Copy(client, bytes)
}

/*----------------------------CLIENT----------------------------*/

/**
* send local songs to server,
* tracker other peers now have this ip
 */
func become_discoverable(args []string) {
	songs := get_local_song_info(args[2])
	msg_content := ""
	for _, s := range songs {
		msg_content += s
	}
	msg := prepare_msg(INIT, 0, []byte(msg_content))
	tracker := send(*msg, TRACKER_IP+args[1])
	defer tracker.Close()
}

func prepare_msg(t byte, id int, content []byte) *TSP_msg {
	return &TSP_msg{TSP_header{t, id}, content}
}

func send(msg TSP_msg, dest_ip string) (conn net.Conn) {
	conn, err := net.Dial("tcp", dest_ip)
	if err != nil {
		fmt.Println("error connecting to " + dest_ip)
		os.Exit(1)
	}
	encoder := gob.NewEncoder(conn)
	encoder.Encode(msg)
	return
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
* prints master list received from tracker
 */
func print_master_list(list string) {
	rows := strings.Split(list, "\n")
	for _, r := range rows {
		song_id := strings.Split(r, ":")[0]
		song_name := strings.Split(r, ",")[1]
		song_artist := strings.Split(r, ",")[2]
		end := strings.Index(song_artist, ">")
		fmt.Println(song_id + ":" + song_name + "," + song_artist[:end])
	}
	fmt.Println(" ")
}

func get_cmd() string {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
	query := "Select option"
	cmd, _ := ui.Select(query, []string{"LIST", "INFO", "PLAY", "STOP", "QUIT"}, &input.Options{
		Loop: true,
	})
	return cmd
}

func get_song_info(id string) {
	songs := strings.Split(master_list, "\n")
	for _, s := range songs {
		song_id := strings.Split(s, ":")[0]
		if song_id == id {
			fmt.Println(s + "\n")
			return
		}
	}
	fmt.Println("Song not found.")
	return
}

func get_song_selection() (int, string) {
	songs := strings.Split(master_list, "\n")
	var ip string

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
	query := "Select a song"
	//fmt.Println(songs)
	id, _ := ui.Ask(query, &input.Options{
		ValidateFunc: func(id string) error {
			for _, s := range songs {
				song_id := strings.Split(s, ":")[0]
				if song_id == id {
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

/**
* handle input command from the user
* LIST - get song list from peers
* PLAY <song id> - play song
* PAUSE - pauses playing of song (buffering continues)
* STOP - stop streaming song
* QUIT - <--
 */
func handle_command(args []string, play chan bool, stop chan bool) int {
	cmd := get_cmd()

	switch cmd {
	case "LIST":
		msg := prepare_msg(LIST, 0, nil)
		tracker := send(*msg, TRACKER_IP+args[1])
		receive_master_list(tracker)
	case "PLAY":
		id, peer_ip := get_song_selection()
		msg := prepare_msg(PLAY, id, nil)
		peer := send(*msg, peer_ip+args[1])
		go receive_mp3(peer, play, stop)
		play <- true
	case "INFO":
		id, _ := get_song_selection()
		get_song_info(strconv.Itoa(id))
		//	msg := prepare_msg(INFO, id, nil)

	case "STOP":
		stop <- true
	case "QUIT":
		msg := prepare_msg(QUIT, 0, nil)
		_ = send(*msg, TRACKER_IP+args[1])
		return -1
	default:
		fmt.Println("invalid command")
	}
	return 0
}

// hello
func receive_mp3(server net.Conn, play chan bool, stop chan bool) {
	defer server.Close()
	for {
		select {
		case <-stop:
			return
		case <-play:
			decoder, err := mp3.NewDecoder(server)
			if err != nil && err == io.EOF {
				return
			}
			if err != nil && err != io.EOF {
				panic(err)
			}
			defer decoder.Close()
			player, err := oto.NewPlayer(decoder.SampleRate(), 2, 2, 8192)
			if err != nil && err == io.EOF {
				return
			}
			if err != nil && err != io.EOF {
				panic(err)
			}
			defer player.Close()

			go func() {
				_, _ = io.Copy(player, decoder)
			}()
		}
	}
}
