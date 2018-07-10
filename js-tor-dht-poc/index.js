const CID = require('cids')
const libp2p = require('libp2p')
const DHT = require('libp2p-kad-dht')
const MPLEX = require('libp2p-mplex')
const Bootstrap = require('libp2p-railing')
const SECIO = require('libp2p-secio')
const WebSockets = require('libp2p-websockets')
const PeerInfo = require('peer-info')

if (!window.tordhtpoc) window.tordhtpoc = (function() {
  // Class extending libp2p providing the main functionality
  class Node extends libp2p {
    constructor(peerInfo, bootstrapPeer) {
      const bootstrapPeerAddr = peerToIpfsAddress(bootstrapPeer)
      console.log('Peer from ' + bootstrapPeer + ' to ' + bootstrapPeerAddr)
      // Construct the base
      super({
        peerInfo: peerInfo,
        modules: {
          transport: [
            new WebSockets()
          ],
          streamMuxer: [
            MPLEX
          ],
          connEncryption: [
            SECIO
          ],
          peerDiscovery: [
            Bootstrap
          ],
          dht: DHT
        },
        config: {
          peerDiscovery: {
            bootstrap: {
              interval: 10000,
              list: [
                bootstrapPeerAddr
              ]
            }
          },
          dht: {
            kBucketSize: 20
          },
          EXPERIMENTAL: {
            dht: true
          }
        }
      })
    }
  }

  function peerToIpfsAddress(peer) {
    // Peer is in form like <onion-id>:<port>/<ipfs-id>
    // Need to change to /dns4/<onion-id>.onion/tcp/<port>/ws/ipfs/<ipfs-id>
    const slashIndex = peer.lastIndexOf('/')
    if (slashIndex == -1) throw new Error('No slash')
    const colonIndex = peer.lastIndexOf(':', slashIndex)
    if (colonIndex == -1) throw new Error('No colon')
    return '/dns4/' + peer.substring(0, colonIndex) + '.onion' +
        '/tcp/' + peer.substring(colonIndex + 1, slashIndex) +
        '/ws/ipfs/' + peer.substring(slashIndex + 1)
  }

  function ipfsAddressToPeer(id, addr) {
    // Inverse of above except addr is js-multiaddr form
    const tuples = addr.stringTuples()
    if (tuples.length < 3) throw new Error('Invalid tuple count')
    return tuples[0][1] + ':' + tuples[1][1] + '/' + id
  }

  return {
    debugEnable: (str) => debug.enable(str),
    bodyOnLoad: () => {
      console.log('Attaching DOM handlers')
      document.getElementById('button-find').onclick = () => {
        // Clear out the results
        document.getElementById('find-results').innerHTML = 'Please wait...'
        // Run the find and put new results
        tordhtpoc.findProviders(document.getElementById('text-peer').value, (err, peers) => {
          let newHtml
          if (err) {
            newHtml = 'Error finding peers: ' + err
          } else {
            newHtml = 'Found ' + peers.length + ' peers<br />'
            peers.forEach(peer => newHtml += peer + '<br />' )
          }
          document.getElementById('find-results').innerHTML = newHtml
        })
      }
    },
    findProviders: (bootstrapPeer, origCallback) => {
      console.log('Finding providers starting from bootstrap', bootstrapPeer)
      PeerInfo.create((err, peerInfo) => {
        if (err) return origCallback(err)
        const node = new Node(peerInfo, bootstrapPeer)
        console.log('Created node', node)
        const callback = (err, peers) => {
          if (!origCallback) return
          const cb = origCallback
          origCallback = null
          node.stop(() => {})
          if (err) return cb(err)
          const peerAddrs = peers.map(peer => {
            const addrArray = peer.multiaddrs.toArray()
            if (addrArray.length == 0) return '<unknown-address>/' + peer.id.toB58String()
            return ipfsAddressToPeer(peer.id.toB58String(), peer.multiaddrs.toArray()[0])
          })
          cb(null, peerAddrs)
        }
        
        let firstDiscover = true
        let firstRouteAdd = true
        node.on('peer:discovery', (peerInfo) => {
          console.log('Discovered peer', peerInfo.id._idB58String, peerInfo, node._dht.routingTable.size)
          if (firstDiscover) {
            firstDiscover = false
            
            console.log('Attaching routing table listener')
            node._dht.routingTable.kb.on('added', () => {
              console.log('Peer added to routing table')
              if (firstRouteAdd) {
                firstRouteAdd = false
                console.log('Finding providers')
                node.contentRouting.findProviders(new CID('zSZ2t1xfC1BLUsxYf9p3ZMwdi1rdvpi4nKveRqujSLoMzFGGb'), 60000, (err, peers) => {
                  console.log('Find complete', err, peers)
                  callback(err, peers)
                })
              }
            })

            console.log('Dialing peer')
            node.dial(peerInfo, (err, conn) => {
              if (err) {
                console.log('Dial failure', err)
                return callback(err)
              }
              console.log('Dial success', conn)
            })
          }
        })
        
        node.on('peer:connect', (peerInfo) => {
          console.log('Connected peer', peerInfo, node._dht.routingTable.size)
        })

        node.on('peer:disconnect', (peerInfo) => {
          console.log('Disconnected peer', peerInfo)
        })
        
        console.log('Starting node')
        node.start((err) => {
          if (err) return callback(err)
          console.log('Node started')
        })
      })
    }
  }
})()