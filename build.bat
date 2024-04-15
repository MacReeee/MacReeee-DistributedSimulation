protoc --go_out . --go_opt=paths=source_relative --go-grpc_out . --go-grpc_opt=paths=source_relative ./hotstuff/pb/proto/*.proto

go build -v -o ./console.exe ./cmd/hotstuff/console/main.go

go build -v -o ./hotstuff ./cmd/hotstuff/server/main.go

docker run -dit \
  --name console \
  --network=HotStuff \
  -v /mnt/c/Users/Yang/Desktop/distributedHotstuff:/home/hotstuff/distributedHotstuff \
  --user=hotstuff \
  --workdir=/home/hotstuff/distributedHotstuff \
  hotstuff:2 \
  /bin/bash -c "tail -f /dev/null"

docker run -dit  --name console --network=HotStuff -v ./:/home/hotstuff/distributedHotstuff --user=hotstuff --workdir=/home/hotstuff/distributedHotstuff hotstuff:2 /bin/bash -c "tail -f /dev/null"


docker exec -it console /bin/bash

tmux attach -t console