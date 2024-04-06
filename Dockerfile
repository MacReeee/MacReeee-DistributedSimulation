FROM ubuntu:20.04

LABEL authors="Yang"
WORKDIR /app

COPY ./blockchain .
COPY ./cmd .
COPY ./consensus .
COPY ./cryp .
COPY ./middleware .
COPY ./modules .
COPY ./pb .
COPY ./view .

RUN go mod download
WORKDIR /app/cmd/hotstuff/server
RUN go build -o hotstuff .
CMD ["./hotstuff"]
