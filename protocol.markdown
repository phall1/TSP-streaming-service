# Torero Streaming Protocol (TSP)
---
### Abstract

The Torero Streaming Protocol, or TSP, is an application-level protocol for
contol over the delivery of data with real-time properties. TSP provides a
framework to enable controlled, on-demand deliver of real-time data, such as
audio files. Sources of this data must be stored clips, but can be extended to
include live feeds.

### Something 
---

#### Tracker Server 

The tracker server keeps track of the all peers currently running the program, and the 
mp3 files that they host. When starting the program, peers first  initiate 
a connection with the tracker, to let the tracker know they are on 
the network. The tracker then requests a list of all songs that the peer is 
hosting, and adds them to the list hosted by the tracker. 

* 'list' 
    * replies with a list of songs, and the machines on which they are hosted
* 'info'
    * provides info for the song requested
    * returns this to the client

#### Peers
##### Outgoing messages 
* 'list' 
    * Requests a list of songs from the tracker
    * Tracker returns list of songs and their associated ips
        * ?? Should we keep this list for when we want to play??
* 'info' 
    * Requests other info for the song from the 
* 'play'
    * requests ip address of peer hosting the specified song
    * streams the song from the appropriate client
* 'stop' - stops playing and closes connection with peer if not yet closed

##### Incoming messages 
* ''


### Protocol - RTSP
---
RTSP has a state
* ID is used when needed to track concurrent sessions??

* Uses TCP to maintain end-to-end connection
