FROM golang
COPY . ./app
RUN cd app; go install github.com/piokra/macedonio/backend; go get -u; go build .
CMD cd app; go run .