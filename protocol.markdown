# Torero Streaming Protocol (TSP)
---
### Abstract

The Torero Streaming Protocol, or TSP, is an application-level protocol for
contol over the delivery of data with real-time properties. TSP provides a
framework to enable controlled, on-demand deliver of real-time data, such as
audio files. Sources of this data must be stored clips, but can be extended to
include live feeds.

IMPORTANT: Our project will be written entirely in go (golang).

### Header Format
---

The messages will be packaged with a standard TCP header.
TSP information will be located at the beginning of the TCP data payload.
Our header will contain:
| Request Type (1 byte) | Song ID (4 byte int) |
|:---------------------:|:--------------------:|
This header could be followed by encoded mp3 data if necessary.

#### Tracker Server 

The tracker server keeps track of all the peers currently running the program, and the 
mp3 files that they host. When starting the program, peers first initiate 
a connection with the tracker, to let the tracker know they are on 
the network. The tracker then requests a list of all songs that the peer is 
hosting, and adds them to the list hosted by the tracker. 

##### Incoming messages
* `list` 
    * replies with a list of songs, and the machines on which they are hosted
* `info <song id>`
    * provides info for the song requested
    * returns this to the client
    * if an id is specified, only send info for that song
    * no additional args sends info for all songs
##### Outgoing messages
* `list.info`
    * sends list of available songs and their associated hosts
* `song.info`
    * sends information about a particular song
* `broadcast_list.info`
    * sends list of available songs and their associated hosts to all peers

#### Peers

##### Outgoing messages
* `list` 
    * Requests a list of songs from the tracker
    * Tracker returns list of songs and their associated ips
        * ?? Should we keep this list for when we want to play??
* `info` 
    * Requests other info for the song from the tracker
* `play`
    * requests ip address of peer hosting the specified song
    * streams the song from the appropriate client
* `stop` 
    * stops playing and closes connection with peer if not yet closed

##### Incoming messages 
* `info`
    * sends associated song data to the requester
* `play`
    * sends the requested mp3 file to the requester
* `stop`
    * stops sending data and closes connection
