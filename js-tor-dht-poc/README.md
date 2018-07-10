## JS + Tor + DHT Proof of Concept

This is a proof of concept showing how to have JS connect and find providers from a DHT built in Go. This is
accomplished creating an onion service via [bine](https://github.com/cretz/bine) and have the JS use websockets via
[js-libp2p](https://github.com/libp2p/js-libp2p) to communicate with the [go-tor-dht-poc](../go-tor-dht-poc) impl.

### Running the Service

This assumes the latest nodejs/npm, Go, and Tor are on the `PATH`. First, the website's `index.js` needs to be built
from the dev version. In this directory, after running `npm install`, run `npm run browserify`. This will create a file
at `public/index.js`. It is quite large for this dev version (3.2MB on last check). Rerun this command again if anything
in `index.js` is changed.

Now that the JS file is there, build and run the onion service website hoster via `go build && js-tor-dht-poc`. This
will output the address of an onion service to open in the Tor browser and will remain open until enter is pressed.

### Using the Service

Open the Tor browser and navigate to the given URL. Note, this can take a while because the JS is so large. Assuming
that the [go-tor-dht-poc](../go-tor-dht-poc) `provide` command is running, grab any of the peer strings and enter it
into the text box on the webpage. Then click `Find Providers`. It can take a little bit, but if there are no errors, the
first and last peers from the `provide` command (the ones providing the value we're testing) will be listed on the
webpage.

### How it Works

First, the Go code just creates a simple file-system web server that serves the `public/` directory. This web server is
hosted over an onion service and the address is given.

As for the webpage, it is a very simple HTML file that references a JS file. The JS file uses js-libp2p and when find
provider is clicked, it creates a libp2p node and bootstraps it with that one address. Then, once it has confirmed it
has connected and it has been added to the peer list, it issues a find provider which grabs other peers as necessary and
eventually the answer.

### Notes

* The browserified JS is huge at 3.2MB. This is not really acceptable, and it shrunk by over half with Uglify but there
  were other problems with Uglify. I need to try Webpack. But even > 1MB is too big. More investigation needed here.
* I changed the Go side's libp2p onion address format `/onion/<onion-id>:<port>` to `/dns4/<onion-id>/tcp/<port>` since
  that is more easily recognized in libp2p as a normal web address in the Tor browser. Also, the JS side of multiaddr
  doesn't support custom protocols very well as of this writing.
* It takes a little bit for first bootstrapping the DHT, but it's not that bad. I wonder if I remove the secio stuff if
  the time would improve a lot.
* We don't reuse the node on the DHT, we recreate it every time, but ideally you would connect once at the beginning and
  reuse it.
* It is a bit convoluted to get the DHT ready, see [this issue](https://github.com/libp2p/js-libp2p/issues/220) where I
  explain some of the things I had to do.
* Web sockets in the Tor browser work well.