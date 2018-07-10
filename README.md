# Tor DHT PoC

This is a proof of concept to show that we can advertise onion services in Go and discover them in Go or JS.

**Goals**

* In Go, all network interaction over Tor
  * Use existing tested DHT code (e.g. BT mainline or IPFS)
  * Broadcast different onion addresses for a key
  * Discover different onion addresses for a key
* In browser (all network interaction doesn't have to be in Tor since it will run in Tor Browser, so no WebRTC)
  * Use existing tested DHT code (e.g. BT mainline or IPFS)
  * Discover different onion addresses for a key

Probably choosing IPFS here due to WebTorrent's requirement on WebRTC which Tor Browser doesn't implement.

**Results**

* Go - SUCCESS!! See the [go-tor-dht-poc/](go-tor-dht-poc)'s README.md for details
* JS - SUCCESS!! See the [js-tor-dht-poc/](js-tor-dht-poc)'s README.md for details

**Usage Quick Overview**

Ok, so now that we have success in both places, we can show a quick overview of how to use it. This assumes that
`GOPATH` is set with this repo properly checked out in it and that latest stable versions of Go, Tor, and nodejs/npm are
installed and on the `PATH`.

In this test, we create 5 nodes, say that 2 of them provide a certain type of service, and in both the Tor browser and
Go command line, we determine those 2 from the DHT starting with any of the 5 addresses. All over Tor and anonymous.

First, provide the 5 nodes. Build and run the executable to create the DHT and "provide" the services:

    go build && go-tor-dht-poc provide

This will take a sec to build the DHT. The output (sans some Tor warnings):

    2018/07/10 16:54:32 Created peer #1: gvumv3dlefinroaeibjlrsknedmpd73stlqey4552boeqqfrcktkdvqd:52516/QmRdDcR647RjC7UkdZXiodAQoXQuqjiiR9wXas4RBhw71x
    2018/07/10 16:54:49 Created peer #2: vn5sgarqtdmzl27cnbrsfchkmrwm5erjb6jtmcwqras6qmhpmf6mahqd:52519/QmY9tHQMGUVg7xBnLiKGN81J9DCrns5KS4jK8wn8CwqHym
    2018/07/10 16:55:11 Created peer #3: my4f2m3yggnnpoeacjgwie5gc5yxeshzgvu5eukx7lodwgmwf2ek52yd:52523/QmRLFFGsVv6uq1WbHhZjA2gknRUwQTyLGdiTAs344egWp8
    2018/07/10 16:55:25 Created peer #4: q2fdsiu7eyoj42iii5o44rnrhxncnljexu6zncw44bpdk3udoveyi6ad:52528/QmNQanzjr862heY5ndTuMQPhbEa8GfyniVk6dhkoGXb5MS
    2018/07/10 16:55:35 Created peer #5: lhcaimk5pvkj2p3ewcdrunryout2wgpjasfum45rochkxwi2y4raauyd:52535/QmSMd8DytSqkAJunHfbG4phedaDs3ugmGFMZY3e3UyePJG
    2018/07/10 16:55:35 Providing key on the first one (gvumv3dlefinroaeibjlrsknedmpd73stlqey4552boeqqfrcktkdvqd:52516/QmRdDcR647RjC7UkdZXiodAQoXQuqjiiR9wXas4RBhw71x)
    2018/07/10 16:55:37 Providing key on the last one (lhcaimk5pvkj2p3ewcdrunryout2wgpjasfum45rochkxwi2y4raauyd:52535/QmSMd8DytSqkAJunHfbG4phedaDs3ugmGFMZY3e3UyePJG)
    2018/07/10 16:55:42 Press enter to quit...

So the 5 nodes were created and we are marking #1 and #5 as "providers". Now, we can connect to this DHT through any of
the nodes and find that #1 and #5 are the providers. With this running, in a new console issue a "find" and provide,
say, peer #2's address (don't need to build, already did above):

    go-tor-dht-poc find vn5sgarqtdmzl27cnbrsfchkmrwm5erjb6jtmcwqras6qmhpmf6mahqd:52519/QmY9tHQMGUVg7xBnLiKGN81J9DCrns5KS4jK8wn8CwqHym

After a sec, here's the output:

    2018/07/10 16:58:17 Found data ID on gvumv3dlefinroaeibjlrsknedmpd73stlqey4552boeqqfrcktkdvqd:52516/QmRdDcR647RjC7UkdZXiodAQoXQuqjiiR9wXas4RBhw71x
    2018/07/10 16:58:17 Found data ID on lhcaimk5pvkj2p3ewcdrunryout2wgpjasfum45rochkxwi2y4raauyd:52535/QmSMd8DytSqkAJunHfbG4phedaDs3ugmGFMZY3e3UyePJG

Now let's connect to this DHT from the Tor browser and find them that way. With the DHT "provide" still running,
navigate to `js-tor-dht-poc`. Run `npm install` to install dependencies and to prepare `index.js` for use at
`public/index.js`, run `npm browserify`. This makes a fairly large JS file that our page uses. Now, to start the web
server, build and run the executable via:

    go build && js-tor-dht-poc

The output will be something like:

    Open Tor browser and navigate to http://n6ls3ltbmoyv2ucblvwz7skjg6mmajt6zfi33skwbhanomgexye6r5ad.onion
    Press enter to exit

Now take that address and open it in the Tor browser (it may take a bit for the large JS file to load). Now take any
peer address, say, peer #4 this time and put it in the text box and click `Find Providers. The result after a some time
should be similar to:

    Found 2 peers
    gvumv3dlefinroaeibjlrsknedmpd73stlqey4552boeqqfrcktkdvqd.onion:52516/QmRdDcR647RjC7UkdZXiodAQoXQuqjiiR9wXas4RBhw71x
    lhcaimk5pvkj2p3ewcdrunryout2wgpjasfum45rochkxwi2y4raauyd.onion:52535/QmSMd8DytSqkAJunHfbG4phedaDs3ugmGFMZY3e3UyePJG

Yay, you have an anonymous DHT that can be joined by other programs and have pieces discovered via other programs or via
a Tor-enabled browser.