# Contributing

Thanks for taking the time to join our community and start contributing!

These guidelines will help you get started with the project.

## Building from source

This section describes how to build the application from source.

### Prerequisites

1. *Install Go*

    This application requires [Go 1.9][1] or later.
    We also assume that you're familiar with Go's [`GOPATH` workspace][3]
    convention, and have the appropriate environment variables set.

2. *Install `dep`*

    We use [`dep`][2] for dependency management. `dep` is a fast moving
    project so even if you have installed it previously, it's a good idea to
    update to the latest version using the `go get -u` flag.

        $ go get -u github.com/golang/dep/cmd/dep

### Fetch the source

We use [`dep`][2] for dependency management, but to reduce the size of the
repository, does not include a copy of its dependencies. This might change in
the future, but for now use the following command to fetch the source of the
application and its dependencies:

    $ go get -d github.com/dcberg/kubernetes-sysdig-metrics-apiserver
    $ cd $GOPATH/src/github.com/dcberg/kubernetes-sysdig-metrics-apiserver
    $ dep ensure -vendor-only

Go is very particular when it comes to the location of the source code in your
`$GOPATH`. The easiest way to make the `go` tool happy is to rename the
appliations 's remote location to something else, and substitute your fork for
`origin`. For example, to set `origin` to your fork, run this command
substituting your GitHub username where appropriate.

    $ git remote rename origin upstream
    $ git remote add origin git@github.com:foobar/k8s-sysdig-adapter.git

This ensures that the source code on disk remains at
`$GOPATH/src/github.com/dcberg/kubernetes-sysdig-metrics-apiserver` while the remote repository
is configured for your fork.

The remainder of this document assumes your terminal's working directory is
`$GOPATH/src/github.com/dcberg/kubernetes-sysdig-metrics-apiserver`.

### Building

To build the application, run:

    $ go build ./cmd/adapter

This assumes your working directory is set to
`$GOPATH/src/github.com/dcberg/kubernetes-sysdig-metrics-apiserver`. If you're somewhere else in
the file system you can instead run:

    $ go build github.com/dcberg/kubernetes-sysdig-metrics-apiserver/cmd/adapter

This produces a `adapter` binary in your current working directory.

_TIP_: You may prefer to use `go install` rather than `go build` to cache build
artifacts and reduce future compile times. In this case the binary is placed in
`$GOPATH/bin/adapter`.

### Running the unit tests

Once you have the application building, you can run all the unit tests for the
project:

    $ go test ./...

This assumes your working directory is set to
`$GOPATH/src/github.com/dcberg/kubernetes-sysdig-metrics-apiserver`. If you're working from a
different directory, you can instead run:

    $ go test github.com/dcberg/kubernetes-sysdig-metrics-apiserver/...

To run the tests for a single package, change to package directory and run:

    $ go test .

_TIP_: If you are running the tests often, you can run
`go test -i github.com/dcberg/kubernetes-sysdig-metrics-apiserver/...` occasionally to reduce
test compilation times.

## Contribution workflow

This section describes the process for contributing a bug fix or new feature.
It follows from the previous section, so if you haven't set up your Go workspace
and built the application from source, do that first.

### Before you submit a pull request

This project operates according to the _talk, then code_ rule.
If you plan to submit a pull request for anything more than a typo or obvious
bug fix, first you _should_ [raise an issue][4] to discuss your proposal,
before submitting any code.

### Pre commit CI

Before a change is submitted it should pass all the pre commit CI jobs.
If there are unrelated test failures the change can be merged so long as a
reference to an issue that tracks the test failures is provided.

Once a change lands in master it will be built and available at this tag:
`sevein/k8s-sysdig-adapter:master`. You can read more about the available images
in the [tagging][5] document.

### Build an image

To build an image of your change using the applications's `Dockerfile`, run
these commands (replacing the repository host and tag with your own):

    $ docker build -t docker.io/yourusername/k8s-sysdig-adapter:latest .
    $ docker push docker.io/yourusername/k8s-sysdig-adapter:latest

### Verify your change

Deploy the image you've built to verify the change. More instructions on this
regards will be added soon.

[1]: https://golang.org/dl/
[2]: https://github.com/golang/dep
[3]: https://golang.org/doc/code.html
[4]: https://github.com/dcberg/kubernetes-sysdig-metrics-apiserver/issues
[5]: docs/tagging.md
