# KDAG
## BFT Consensus platform for distributed applications.

[![CircleCI](https://circleci.com/gh/mosaicnetworks/babble.svg?style=svg)](https://circleci.com/gh/mosaicnetworks/babble)

Kdag allows many computers to behave as one. It uses Peer to Peer (P2P) 
networking and a consensus algorithm to guarantee that multiple connected 
computers process the same commands in the same order; a technique known as 
state machine replication. This makes for secure systems that can tolerate 
arbitrary failures including malicious behavior.

For guidance on how to install and use kdag please visit our 
[documentation](http://docs.kdag.io) pages.

**NOTE**:
This is alpha software. Please contact us if you intend to run it in production.

## Consensus Algorithm and Blockchain

We use an adaptation of the Hashgraph consensus algorithm, invented by Leemon 
Baird. Hashgraph is best described in the 
[white-paper](http://www.swirlds.com/downloads/SWIRLDS-TR-2016-01.pdf) and its 
[accompanying document](http://www.swirlds.com/downloads/SWIRLDS-TR-2016-02.pdf).
The original algorithm is protected by [patents](http://www.swirlds.com/ip/) in 
the USA. Therefore, anyone intending to use this software in the USA should 
obtain a license from the patent holders.

Hashgraph is based on the intuitive idea that gossiping about gossip itself 
yields enough information to compute a consensus ordering of events. It attains 
the theoretical limit of tolerating up to one-third of faulty nodes without 
compromising on speed. For those familiar with the jargon, it is a leaderless, 
asynchronous BFT consensus algorithm.

Kdag projects the output of the consensus algorithm onto a linear blockchain 
which is more suitable for representing an ordered list of transactions and 
facilitates the creation of light-clients. For information about this projection
please refer to [documentation](http://docs.kdag.io/blockchain.html) pages.

## Design

Kdag is designed to integrate with applications written in any programming 
language.

### Overview

```
    +--------------------------------------+
    | Client App                                  |
    |                                      |
    |  +-------------+     +------------+  | 
    |  | Service     | <-- | State      |  |
    |  |             |     |            |  |
    |  +-------------+     +------------+  |
    |          |                ^          |
    |          |                |          |
    +----------|----------------|----------+
               |                |                      
--------- SubmitTx(tx) ---- CommitBlock(Block) ------- JSON-RPC/TCP or in-memory       
               |                |                         
 +-------------|----------------|------------------------------+
 | KDAG      |                |                              |
 |             v                |                              |
 |          +----------------------+                           |
 |          | App Gateway            |                           |
 |          |                      |                           |
 |          +----------------------+                           |
 |                     |                                       |
 |   +-------------------------------------+                   |
 |   | Core                                |                   |
 |   |                                     |                   |
 |   |  +------------+                     |    +----------+   |
 |   |  | Hashgraph  |       +---------+   |    | Service  |   |
 |   |  +------------+       | Store   |   | -- |          | <----> HTTP 
 |   |  +------------+       +----------   |    |          |   |
 |   |  | Blockchain |                     |    +----------+   |
 |   |  +------------+                     |                   |
 |   |                                     |                   |
 |   +-------------------------------------+                   |
 |                     |                                       |
 |   +-------------------------------------+                   |
 |   | Edge interface                      |                   |
 |   |                                     |                   |
 |   +-------------------------------------+                   |
 |                     ^                                       |
 +---------------------|---------------------------------------+
                       |
                       v
                  P2P Network
```

Almost any software application can be modeled in terms of a *service* and a 
*state*. The *service* is responsible for processing commands (ex. user input), 
while the *state* is responsible for manipulating and storing the data (eg. 
database). Usually, when commands need to update the data, the *service* will 
invoke the *state* directly. In a distributed application, however, commands 
(referred to as *transactions* in this context), must be broadcast to all 
replicas and consistently ordered before being applied to the *state*. This 
ensures that all replicas process the same commands in the same order. Hence, 
the *service* no longer communicates directly with the *state* (except for 
read-only requests), but forwards commands to a *transaction ordering system* 
which takes care of broadcasting and ordering the transactions across all 
replicas before feeding them back to the application's *state*. 

Kdag is an ordering system that plugs into any application thanks to a very 
simple interface. It uses a consensus algorithm, to replicate and order the 
transactions, and a blockchain to represent the resulting list. A blockchain is 
a linear data structure composed of batches of transactions, hashed and signed 
together, easily allowing to verify any transaction. So, instead of applying 
commands directly to the *state*, Kdag applications must forward the commands 
to Kdag and let them be processed asynchronously by the consensus system 
before receiving them back, in blocks, ready to be applied to the *state*.  

### API

Kdag communicates with the App through an `AppProxy` interface, which has two
implementations:

- `SocketProxy`: A SocketProxy connects to an App via TCP sockets. It enables 
                 the application to run in a separate process or machine, and to 
                 be written in any programming language.

- `InmemProxy` : An InmemProxy uses native callback handlers to integrate Kdag 
                 as a regular Go dependency. 

The `AppProxy` interface exposes three methods for Kdag to call the App:

- `CommitBlock(Block) ([]byte, error)`: Commits a block to the application and 
                                        returns the resulting state hash.
- `GetSnapshot(int) ([]byte, error)`: Gets the application snapshot 
                                      corresponding to a particular block index.
- `Restore([]byte) error`: Restores the App state from a snapshot.

Reciprocally, `AppProxy` relays transactions from the App to Kdag via a native 
Go channel - `SubmitCh` - which ties into the application differently depending 
on the type of proxy (Socket or Inmem).

kdag asynchronously processes transactions and eventually feeds them back to 
the App, in consensus order and bundled into blocks, with a **CommitBlock** 
call. Transactions are just raw bytes and Kdag does not need to know what they 
represent. Therefore, encoding and decoding transactions is done by the App.

Refer to the [API documentation](http://docs.kdag.io/latest/api.html) for more 
details and a quick-start guide.

## Dev

### Go
Kdag is written in [Golang](https://golang.org/). Hence, the first step is to 
install **Go version 1.9 or above** which is both the programming language and a 
CLI tool for managing Go code. Go is very opinionated and will require you to 
[define a workspace](https://golang.org/doc/code.html#Workspaces) where all your 
go code will reside.

### Kdag and dependencies
Clone the [repository](https://github.com/Kdag-K/kdag) in the appropriate 
GOPATH subdirectory:

```bash
$ mkdir -p $GOPATH/src/github.com/Kdag-K/
$ cd $GOPATH/src/github.com/Kdag-K
[...]/mosaicnetworks$ git clone https://github.com/Kdag-K/kdag.git
```
Kdag uses [Glide](http://github.com/Masterminds/glide) to manage dependencies.

```bash
[...]/kdag$ curl https://glide.sh/get | sh
[...]/kdag$ glide install
```
This will download all dependencies and put them in the **vendor** folder.

### Other requirements

Bash scripts used in this project assume the use of GNU versions of coreutils.
Please ensure you have GNU versions of these programs installed:-

example for macos:
```
# --with-default-names makes the `sed` and `awk` commands default to gnu sed and gnu awk respectively.
brew install gnu-sed gawk --with-default-names
```
