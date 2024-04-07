protoc --go_out . --go_opt=paths=source_relative --go-grpc_out . --go-grpc_opt=paths=source_relative ./hotstuff/pb/proto/*.proto

docker run -dit \
  --name console \
  --network=HotStuff \
  -v /mnt/c/Users/Yang/Desktop/distributedHotstuff:/home/hotstuff/distributedHotstuff \
  --user=hotstuff \
  --workdir=/home/hotstuff/distributedHotstuff \
  hotstuff:2 \
  /bin/bash -c "tail -f /dev/null"


docker exec -it console /bin/bash