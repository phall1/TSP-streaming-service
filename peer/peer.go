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
    args := check_args()

    songs := get_songs(args[2])

    // Connect to tacker
    tracker, err := net.Dial("tcp", "localhost:" + args[1])
    if err != nil {
        fmt.Println("some error")
    }

    send_song_list(tracker, songs)

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

func check_args() []string{
    args := os.Args[:]
    if len(args) != 3 {
        fmt.Println("Usage: ", args[0], "<port> <filedir>")
        os.Exit(1)
    }
    return args
}

/*
 * Returns an array of FileInfo types
 * for all files(songs) in directory "songs"
 */
func get_songs(dir_name string) []os.FileInfo {

    songs, err := ioutil.ReadDir(dir_name)
    if err != nil {
        fmt.Println("cant read songs")
        os.Exit(1)
    }

    for _, f := range songs {
        if path.Ext(f.Name()) != ".mp3" {
            fmt.Println(f.Name())
        }
    }
    return songs
}

/*
 * send this peers song list to the 
 * tracker server
 */
func send_song_list(tracker net.Conn, songs []os.FileInfo) {
    
}

