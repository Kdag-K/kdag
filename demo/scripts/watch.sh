#!/bin/bash

# docker run -it --rm --name=watcher --net=kdagnet --ip=172.77.5.99 Kdag-K/watcher /watch.sh $N  

watch -t -n 1 '
docker ps -aqf name=node | \
xargs docker inspect -f "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}" | \
xargs -I % curl -s -m 1 http://%:80/stats | \
tr -d "{}\"" | \w
awk -F "," '"'"'{gsub (/[,]/," "); print;}'"'"'
'
