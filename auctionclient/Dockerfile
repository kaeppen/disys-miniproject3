FROM golang:latest

WORKDIR /app

RUN export GO111MODULE=on

COPY go.mod ./ 
COPY go.sum ./ 
RUN go mod download 
COPY *.go ./

RUN cd /app && git clone https://github.com/kaeppen/disys-miniproject3.git
RUN cd /app/disys-miniproject3/auctionclient && go build -o main . 



CMD ["/app/disys-miniproject3/auctionclient/main"]