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

    songs := get_local_songs(args[2])

    // Connect to tacker
    tracker, err := net.Dial("tcp", "localhost:" + args[1])
    if err != nil {
        fmt.Println("some error")
    }

    send_local_songs(tracker, songs)

    // receive remote song list

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

/*
 * Returns an string slice
 * for all files(songs) in directory "songs"
 */
// func get_songs(dir_name string) []os.FileInfo {
func get_local_songs(dir_name string) []string {
    songs, err := ioutil.ReadDir(dir_name)
    if err != nil {
        fmt.Println("cant read songs")
        os.Exit(1)
    }

    songs_slice := make([]string, 0)
    for i := 0; i < len(songs); i++ {
        if path.Ext(songs[i].Name()) == ".mp3" {
            songs_slice = append(songs_slice, songs[i].Name())
        }
    }
    return songs_slice 
}

/*
 * send this peers song list to the 
 * tracker server
 */
// func send_song_list(tracker net.Conn, songs []os.FileInfo) {
func send_local_songs(tracker net.Conn, songs []string) {
    msg := ""
    for _, s := range songs {
        msg += s + "\n"
    }
    client.Write([]byte(msg))
}

