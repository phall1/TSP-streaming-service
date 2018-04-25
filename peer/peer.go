package main

import (
    "os"
    "net"
    "fmt"
    "bufio"
    "path"
    "io/ioutil"
)

func main() {
    args := os.Args[:]
    if len(args) != 3 {
        fmt.Println("Usage: ", args[0], "<port> <filedir>")
        os.Exit(1)
    }

    songs := get_local_song_info(args[2])

    // Connect to tacker
    tracker, err := net.Dial("tcp", "localhost:" + args[1])
    if err != nil {
        fmt.Println("some error")
    }

    send_local_song_info(tracker, songs)

    // receive remote song list ## STORE IN MEMORY ##


    // wait for user action (play, other stuff)

    for {
        reader := bufio.NewReader(os.Stdin)
        fmt.Println("text to send: ")
        text, _ := reader.ReadString('\n')

        // send to Socket
        fmt.Fprintf(tracker, text + "\n")

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
            continue;
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
    msg := ""
    for _, s := range songs {
        msg += s 
    }
    tracker.Write([]byte(msg))
}

