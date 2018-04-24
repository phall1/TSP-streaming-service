package main

import (
    "net"
    "os"
    "fmt"
    // "io/ioutil"

    // "github.com/gorilla/websocket"
)

const BACKLOG = 10
const MAX_EVENTS = 64

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

    // setup server socket
    ln, err := net.Listen("tcp", ":"+args[1])
    if err != nil {
        fmt.Println(err)
        fmt.Println("error setting up socket")
        os.Exit(1)
    }

    for {
        conn, err := ln.Accept()
        if err != nil {
            // error accepting this connection
            continue
        }
        //fmt.Println(conn)
        go handleConnection(conn)
    }
}

func handleConnection(client net.Conn) {
    buffer := make([]byte, 4096)
    client.Read(buffer)

    fmt.Printf(buffer)
    // fmt.Println(client)
    // testmsg := "hello"
    // client.Write([]byte(testmsg + "\n"))
    // get songs from client
    // send out info about songs to all hosts


    // receive songs from client

    // update info doc (locks)

    // send client (all?) updated doc
}
