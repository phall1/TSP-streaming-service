package main

import (
    "os"
    "net"
    "fmt"
    "bufio"
)


func main() {
    args := os.Args[:]
    if len(args) != 3 {
        fmt.Println("Usage: ", args[0], "<port> <filedir>")
        os.Exit(1)
    }

    // Exit if filedir does not exist
    if _, err := os.Stat(args[2]); os.IsNotExist(err) {
        fmt.Println("filedir does not exist")
        os.Exit(1)
    }

    conn, err := net.Dial("tcp", "localhost:" + args[1])
    if err != nil {
        fmt.Println("some error")
    }

    for {
        reader := bufio.NewReader(os.Stdin)
        fmt.Println("text to send: ")
        text, _ := reader.ReadString('\n')

        // send to Socket
        fmt.Fprintf(conn, text + "\n")

        reply, _ := bufio.NewReader(conn).ReadString('\n')
        fmt.Print("Message from server: " + reply)
    }
}
