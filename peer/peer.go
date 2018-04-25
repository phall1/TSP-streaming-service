package main

import (
	"bufio"
	// "bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
)

type tsp_header struct {
	//one byte for type
	//int song ID
	t       byte
	song_id int
}

type tsp_msg struct {
	header tsp_header
	msg    []byte
}

func main() {
	args := os.Args[:]
	if len(args) != 3 {
		fmt.Println("Usage: ", args[0], "<port> <filedir>")
		os.Exit(1)
	}

	songs := get_local_song_info(args[2])

	// Connect to tacker
	tracker, err := net.Dial("tcp", "localhost:"+args[1])
	if err != nil {
		fmt.Println("some error")
	}

	send_local_song_info(tracker, songs)

	// Close connection with the tracker

	// SPIN

	// receive remote song list ## STORE IN MEMORY ##

	// wait for user action (play, other stuff)

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("text to send: ")
		text, _ := reader.ReadString('\n')

		// send to Socket
		fmt.Fprintf(tracker, text+"\n")

		reply, _ := bufio.NewReader(tracker).ReadString('\n')
		fmt.Print("Message from server: " + reply)
	}
}

func receive_remote_song_info() {

}

/*
 * Returns an string slice
 * for all files(songs) in directory "songs"
 */
// func get_songs(dir_name string) []os.FileInfo {
func get_local_song_info(dir_name string) []string {
	// Read files from directory
	info_files, err := ioutil.ReadDir(dir_name)
	if err != nil {
		fmt.Println("cant read songs")
		os.Exit(1)
	}

	// Extract content of files and add to slice of strings
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

/*
 * send this peers song list to the
 * tracker server
 */
// func send_song_list(tracker net.Conn, songs []os.FileInfo) {
func send_local_song_info(tracker net.Conn, songs []string) {
	msg_content := ""
	for _, s := range songs {
		msg_content += s
	}

	// bin_buff := new(bytes.Buffer)

	encoder := gob.NewEncoder(tracker)

	msg_struct := &tsp_msg{
		header: tsp_header{t: 0, song_id: 0},
		msg:    []byte(msg_content)}

	// fmt.Println(msg_struct)

	encoder.Encode(msg_struct)
	fmt.Println(msg_struct)
	// encoder.Flush()
	tracker.Close()
	//slice := hdr
	//fmt.Printf("%T\n", hdr)
	// send_buff := append(bin_buff.Bytes(), []byte(msg))
	//slice = append(hdr, []byte(msg))
	// tracker.Write(bin_buff.Bytes())
	// tracker.Write(bin_buff.Bytrs())
	os.Exit(1)
	// tracker.Write(send_buff)
}
