#hvordan fortæller vi den, at den skal builde hhv. klient og server? det skal vel være forskellige images? 
FROM golang:latest

WORKDIR /app

RUN export GO111MODULE=on

COPY go.mod ./ 
COPY go.sum ./ 
RUN go mod download 
COPY *.go ./

RUN cd /app && git clone https://github.com/kaeppen/disys-miniproject3.git
RUN cd /app/disys-miniproject3 && go build -o main . 

ENV PORT $port #er det  overhovedet nødvendigt? 


CMD ["/app/disys-miniproject3/main"]