# drone-s3

[![Build Status](http://beta.drone.io/api/badges/drone-plugins/drone-s3/status.svg)](http://beta.drone.io/drone-plugins/drone-s3)
[![Go Doc](https://godoc.org/github.com/drone-plugins/drone-s3?status.svg)](http://godoc.org/github.com/drone-plugins/drone-s3)
[![Go Report](https://goreportcard.com/badge/github.com/drone-plugins/drone-s3)](https://goreportcard.com/report/github.com/drone-plugins/drone-s3)
[![Join the chat at https://gitter.im/drone/drone](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/drone/drone)

Drone plugin to publish files and artifacts to Amazon S3 or Minio. For the
usage information and a listing of the available options please take a look at
[the docs](http://plugins.drone.io/drone-plugins/drone-s3/).

## Build

Build the binary with the following commands:

```
go build
go test
```

## Docker

Build the Docker image with the following commands:

```
(deprecated) CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo
docker build --rm=true -t plugins/s3 .
```

Please note incorrectly building the image for the correct x64 linux and with
CGO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/bin/drone-s3' not found or does not exist..
```

## Usage

Execute from the working directory:

```
docker run --rm \
  -e PLUGIN_SOURCE=<source> \
  -e PLUGIN_TARGET=<target> \
  -e PLUGIN_BUCKET=<bucket> \
  -e PLUGIN_CREATEBUCKET="true" \
  -e PLUGIN_CREATEBUCKET="true" \
  -e PLUGIN_GITFLOW_READY="true" \
  -e PLUGIN_APPEND_BRANCH="true" \
  -e DRONE_COMMIT_BRANCH="feature/SomeFeatures" \
  -e AWS_ACCESS_KEY_ID=<token> \
  -e AWS_SECRET_ACCESS_KEY=<secret> \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  plugins/s3 --dry-run
```

**I** have *forked* the ***repository*** to ***add*** some **functionalities** that i *needed*.
I added this YAML field:
| Field Name | Type | Usage |
| ------ | ------ | ------ |
| `create-bucket-if-necessary` | **bool** | *Create* ***S3 Bucket***, if not *existing*
| `append-branch-to-bucket` | **bool** | If `create-bucket-if-necessary` is **true**, *append* the **branch name** *triggered* by the ***CI*** in **bucket name**.
| `prefixstripbranch` | **string** | If `append-branch-to-bucket` is **true**, *remove* **prefix** in ***branch name***.
| `s3-hosting` | **bool** | *Activate* **S3 Hosting** on ***Bucket***.
| `indexdocument` | **string** | ***Index Document*** for **S3 Hosting Configuration**
| `errordocument` | **string** | ***Error Document*** for **S3 Hosting Configuration**

For **React JS** or other ù, do not *forget* to put `index.html` on **error-document**.
It will *redirect* all ***url*** to **index.html**

```
pipeline:
   s3-branch:
     image: lucandjaro/drone-s3:dev
     bucket: bucket-in-s3
     acl: public-read
     source: dist/*
     strip_prefix: dist/
     target: /
     delete: true
     region: eu-west-3
     create-bucket-if-necessary: true
     append-branch-to-bucket: true
     ### If you are using gitflow or any prefix like release/
     prefixstripbranch: "feature/"
     s3-hosting: true
     indexdocument: index.html
     errordocument: index.html
     secrets: [ AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY ]
     when:
      branch: [ feature/* ]
      event: [ push, pull_request ]
```

