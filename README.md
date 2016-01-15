# torro
a BitTorrent client for the purposes of learning Go.

### TODO

#### Bencoding

* Some files are not being parsed correctly. We need to test parser on a larger set of .torrent files.

#### Tracker Protocol

* Support UDP trackers

#### P2P

* Handshake message needs to be sent.
* All other messages need to be sent/received through a state machine.
* Need to design concurrent goroutines for handling clients.

#### File Handling

* Need to write pieces to a memory mapped file.
* If some files are finished but others are not, the finished files should be fully accessible on the filesystem.
* Unfinished files should end in .part extension

#### Database

* Store file/transfer metadata in a central database. sqlite?
* Explore whether to use an ORM.
* Should there be a separate config/settings file?

#### Command Line

* Start a background daemon that can download and upload files
* Commands
    * Start
    * Pause
    * Delete
    * List

#### Console UI

* Explore using [Termbox](https://github.com/nsf/termbox-go) and [Termui](https://github.com/gizak/termui)

#### Web API

* Should this be REST or WebSocket based?

#### Web UI

* React-based UI
* Webpack
* es6 with Babel
* Browser notifications
* d3 Visualizations

### Resources

* [Bittorrent Protocol Specification v1.0](https://wiki.theory.org/BitTorrentSpecification)
* [Kristen Widman - How to Write a Bittorrent Client](http://www.kristenwidman.com/blog/33/how-to-write-a-bittorrent-client-part-1/)

