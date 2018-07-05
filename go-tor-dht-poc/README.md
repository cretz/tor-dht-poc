## Go + Tor + DHT Proof of Concept

This is a proof of concept showing how to build a DHT over Tor with v3 onion services. This was accomplished with the
Tor client library [bine](https://github.com/cretz/bine) and [libp2p](https://github.com/libp2p/go-libp2p)
(specifically the [DHT lib](https://github.com/libp2p/go-libp2p-kad-dht)).

Basically, the goal was to be able to build an onion service, connect to onion service peers, and once a DHT is formed,
be able to signal that some onion services are providers for certain hashes.

### Setup

The executable referenced hencforth as `go-tor-dht-poc` can be built by running `go build` within this directory.

When running the executable, the latest stable `tor` executable needs to be on the `PATH`.

### Running the Provider

The DHT has to be created before anything can be discovered on it. Once built, execute the following:

    go-tor-dht-poc provide

By default this will create 5 onion services and connect them to each other as peers in the DHT. While this is all done
on a single machine, the connections actually occur over the Tor network. Except for a Tor warning or two, here is an
example of the output from running the above:

    2018/07/05 17:53:04 Creating 5 peers
    2018/07/05 17:53:11 Created peer #1: l6dxbz6p7js3zqkmo36pgetvv22jnp5wo5ro5bpphjy2zeckcdzqalad:60236/QmV3nngmTNnbJDT8SfWCvkDzmjQ58grDXuReVb4YshaqA7
    2018/07/05 17:53:24 Created peer #2: 7zj5gwyvueirwo5cj2epllzhfyjiuxcykzwzt72wnbt7oi27wd6plxad:60239/QmWi83L5LVEeEyLpTNhv9ngLczX8WzVzAovygqXeywKA73
    2018/07/05 17:53:41 Created peer #3: teodbqqhaxb7rxavzejrhba7p347twhkzzgjcuhr4zk4slgugbchneqd:60244/QmRdaLUTqtFVZgxRvNvBnK6fU4ygHPXHPkzwxGaJAtFZ5T
    2018/07/05 17:53:54 Created peer #4: 5npciffgygrdggxc7dzhy6j77zf7k3eieiminu6qllz4cnm5kcqjc4id:60249/QmdE5mzfupo8XaKdQCSaQ1DABsVc18rVS2ycHyWkHq33ra
    2018/07/05 17:54:03 Created peer #5: k4ufqglseubr6gmvmgl4ragzumbozgqyytwzrdxunzqnik3idjvvjsyd:60256/QmSQpNVnzVRv7ckQWDQkrd1juzwZ1kaAoutap8AKQUpri7
    2018/07/05 17:54:03 Providing key on the first one (l6dxbz6p7js3zqkmo36pgetvv22jnp5wo5ro5bpphjy2zeckcdzqalad:60236/QmV3nngmTNnbJDT8SfWCvkDzmjQ58grDXuReVb4YshaqA7)
    2018/07/05 17:54:04 Providing key on the last one (k4ufqglseubr6gmvmgl4ragzumbozgqyytwzrdxunzqnik3idjvvjsyd:60256/QmSQpNVnzVRv7ckQWDQkrd1juzwZ1kaAoutap8AKQUpri7)
    2018/07/05 17:54:12 Press enter to quit...

So this created 5 peers (onion services) that are connected in the DHT. Then, for this example, we choose to broadcast
that we have a certain value on the first and the last one. The peer IDs given can then be used for `find` below. Values
can be tweaked in `main.go` including `debug` to see more info.

### Running the Find

The client side of the DHT is very similar except that it doesn't create an onion service, it just connects to peers.
It accepts at least one peer to bootstrap its way on to the DHT. The above example made 5 peers but only broadcasted
that some value was present on two of them. We can take, say, the third peer address as our initial peer into the `find`
call:

    go-tor-dht-poc find teodbqqhaxb7rxavzejrhba7p347twhkzzgjcuhr4zk4slgugbchneqd:60244/QmRdaLUTqtFVZgxRvNvBnK6fU4ygHPXHPkzwxGaJAtFZ5T

The command takes multiple peers, but one is all that is often needed if it's online. The result of the command above
looks like the following (again, with a couple of Tor warnings stripped):

    2018/07/05 17:54:54 Creating DHT and connecting to peers
    2018/07/05 17:55:32 Found data ID on l6dxbz6p7js3zqkmo36pgetvv22jnp5wo5ro5bpphjy2zeckcdzqalad:60236/QmV3nngmTNnbJDT8SfWCvkDzmjQ58grDXuReVb4YshaqA7
    2018/07/05 17:55:32 Found data ID on k4ufqglseubr6gmvmgl4ragzumbozgqyytwzrdxunzqnik3idjvvjsyd:60256/QmSQpNVnzVRv7ckQWDQkrd1juzwZ1kaAoutap8AKQUpri7

This shows the successful demonstration of broadcasting a provider of a certain value on an anonymous DHT.

### How it Works

I will not go in to details about Kademlia DHTs or how peers are routed. This leverages IPFS's DHT because BitTorrent's
DHT has more limitations in the client implementations around info hashes.

To setup the DHT, we created 5 peers. Each peer was given the peers before it during bootstrap and so long as it
connected to a couple I considered it connected. An IPFS transport was implemented over Tor akin to what projects like
[go-onion-transport](https://github.com/OpenBazaar/go-onion-transport/) had done. The existing onion address format in
IPFS has [some limitations](https://github.com/multiformats/multiaddr/issues/65) that kept me from serializing v3
addresses so I just overrode the protocol at runtime. Similarly, I chose to have Tor generate my onion service keys for
me even though I could have easily generated them myself. I created a separate "onionListen" protocol for this to keep
it simple for now.

Once the onion services are set up and connected to one another over Tor, I simply "provide" an ID to the DHT. In this
case I just hashed a hardcoded string. For finding, I give it one of more of the onion addresses and the custom
transport dials them before asking for the peers with that same hash. IPFS assigns peer IDs, so I just chose a simple
string format of onion ID, colon, port, slash, then IPFS peer ID. I could have reused multiaddrs here but I needed it
more generic out of my common interface. 

### Notes

Quick notes:

* Onion multiaddr format is limiting
* Can't "provide" a value for a node that is not yet connected to a peer, which is reasonable of course
* There is something racy in the Tor socks proxy asking for several connections in the same millisecond (or maybe in
  bine code). I did not debug this, I just tossed a 100ms sleep in between requests which mostly solves it.
* The creation/connection of peers is a bit slow at first. The creation is not that bad and v3 onion services are faster
  than v2, but either way they have to upload descriptors to the directory servers and get back successful responses. As
  for why the connection of peers is a bit slow at first is just Tor building circuits via rendezvous points and these
  are ephemeral onion services created just seconds ago, so any caching of service directory entries by relays would
  probably have little effect.
* I have not tried the JS DHT impl yet, but I hope the addr formats and transports are similarly pluggable. I am hoping
  that I can host onion service web sockets from Go and connect to them in the Tor browser to build the DHT there. I
  know the JS impl is very young do we'll see.
* For my use case, this is good enough. I just need to broadcast which onion services say they host something. However,
  some may have other use cases such as mutable data or more complex p2p interactions. There is
  [a libp2p PR](https://github.com/libp2p/go-libp2p/pull/278) that has an ok example of how to set stream handlers. It
  can be combined with the knowledge here to do more p2p-ish things over Tor.