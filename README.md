# Tor DHT PoC

This is a proof of concept to show that we can advertise onion services in Go and discover them in Go or JS.

**Goals**

* In Go, all network interaction over Tor
  * Hook to existing DHT in the wild (e.g. mainline or IPFS)
  * Broadcast different onion addresses for a key
  * Discover different onion addresses for a key
* In browser (all network interaction doesn't have to be in Tor since it will run in Tor Browser, so no WebRTC)
  * Hook to existing DHT in the wild (e.g. mainline or IPFS)
  * Discover different onion addresses for a key

Probably choosing IPFS here due to WebTorrent's requirement on WebRTC which Tor Browser doesn't implement.

**Results**

Nothing yet, still in development