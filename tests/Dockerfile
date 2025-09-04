FROM golang:1.25

COPY . /tests

WORKDIR /tests

ENTRYPOINT [ "go", "test", "-race", "-v", "./..." ]
