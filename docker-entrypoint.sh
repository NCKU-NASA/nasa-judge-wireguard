#!/bin/bash

cp /config/* $WGPATH

for a in $(echo "$SERVER" | jq -r '. | join(" ")')
do
    wg-quick up $a
done

bash init.sh

./nasa-judge-wireguard

