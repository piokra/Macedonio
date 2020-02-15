FROM golang
COPY . /app
RUN GOPATH=$GOPATH:/app
CMD go run /app/main.go