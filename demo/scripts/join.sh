#!/bin/bash

N=${1:-5}
FASTSYNC=${2:-false}
WEBRTC=${3:-false}
DEST=${4:-"$PWD/conf"}

dest=$DEST/node$N

# Create new key-pair and place it in new conf directory
mkdir -p $dest
echo "Generating key pair for node$N"
docker run  \
    -u $(id -u) \
    -v $dest:/.kdag \
    --rm Kdag-K/kdag:latest keygen 

# copy signal TLS certificate
cp $PWD/../src/net/signal/wamp/test_data/cert.pem $dest/cert.pem

# get genesis.peers.json
echo "Fetching peers.genesis.json from node1"
curl -s http://179.23.12.1:80/genesispeers > $dest/peers.genesis.json

# get up-to-date peers.json
echo "Fetching peers.json from node1"
curl -s http://179.23.12.1:80/peers > $dest/peers.json

# start the new node
docker run -d --name=client$N --net=kdagnet --ip=179.23.12.$N -it Kdag-K/dummy:latest \
    --name="client $N" \
    --client-listen="179.23.12.$N:1339" \
    --proxy-connect="179.23.12.$N:1338" \
    --discard \
    --log="debug" 

docker create --name=node$N --net=kdagnet --ip=179.23.12.$N Kdag-K/kdag:latest run \
    --heartbeat=100ms \
    --moniker="node$N" \
    --cache-size=50000 \
    --listen="179.23.12.$N:1337" \
    --proxy-listen="179.23.12.$N:1338" \
    --client-connect="179.23.12.$N:1339" \
    --service-listen="179.23.12.$N:80" \
    --fast-sync=$FASTSYNC \
    --log="debug" \
    --sync-limit=100 \
    --webrtc=$WEBRTC \
    --signal-addr="179.23.12.1:2443"

 # --store \

docker cp $dest node$N:/.kdag
docker start node$N
