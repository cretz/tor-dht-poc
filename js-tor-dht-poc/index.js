
const PeerInfo = require('peer-info')
const libp2p = require('libp2p')
const WebSockets = require('libp2p-websockets')
const SECIO = require('libp2p-secio')
const Bootstrap = require('libp2p-railing')
const DHT = require('libp2p-kad-dht')
const CID = require('cids')
const MPLEX = require('libp2p-mplex')

class Node extends libp2p {
  constructor(peerInfo, bootstrapPeer) {
    console.log('Peer from ' + bootstrapPeer + ' to ' + bootstrapPeerToIpfsAddress(bootstrapPeer))
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
              bootstrapPeerToIpfsAddress(bootstrapPeer)
            ]
          }
        },
        dht: {
          kBucketSize: 20
        },
        EXPERIMENTAL: {
          pubsub: false,
          dht: true
        }
      }
    })
  }
}

function bootstrapPeerToIpfsAddress(peer) {
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

let init = true

console.log('Starting')
PeerInfo.create((err, peerInfo) => {
  if (err) throw err
  const node = new Node(peerInfo, 'srxgfdgzicqozfbv2fdaw26nsfbqfkwqryefszeuu7uhmwwbdxqudxqd:53591/QmXpvxKhn6yEiwHsYNsDb68TmaRNtNRDKjQWq8bnw3rMfX')
  console.log('Created: ', node)
  node.on('peer:discovery', (peerInfo) => {
    const first = init
    init = false
    console.log('Discovered peer', peerInfo.id._idB58String)
    if (first) node.dial(peerInfo, (err, conn) => {
      if (err) { return console.log('Dial failure', err) }
      console.log('Dial success', conn)
      node.contentRouting.findProviders(new CID('zSZ2t1xfC1BLUsxYf9p3ZMwdi1rdvpi4nKveRqujSLoMzFGGb'), 60000, (err, peers) => {
        console.log('Find complete', err, peers)
      })
    })
  })
  node.on('peer:connect', (peerInfo) => {
    console.log('Connected peer', peerInfo)
  })
  node.on('peer:disconnect', (peerInfo) => {
    console.log('Disconnected peer', peerInfo)
  })
  node.start((err) => {
    if (err) throw err
    console.log('Node started')
  })
})