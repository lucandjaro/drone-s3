FROM golang:1.10.3-alpine as compile
COPY . /usr/src/drone-s3/
WORKDIR /usr/src/drone-s3/
RUN apk update
RUN apk add git
RUN go get github.com/Sirupsen/logrus
RUN go get github.com/aws/aws-sdk-go/aws
RUN go get github.com/joho/godotenv
RUN go get github.com/mattn/go-zglob
RUN go get github.com/urfave/cli
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo



FROM centurylink/ca-certs
ENV GODEBUG=netdns=go
ADD contrib/mime.types /etc/
COPY --from=compile /usr/src/drone-s3/drone-s3 /bin/
#ADD release/linux/amd64/drone-s3 /bin/
ENTRYPOINT ["/bin/drone-s3"]
