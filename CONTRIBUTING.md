# Contributing

Thanks for taking the time to join our community and start contributing!

These guidelines will help you get started with the project.

## Building from source

This section describes how to build the application from source.

### Prerequisites

This application requires [Go 1.11][1] or later.

### Fetch the source

With the implementation of the new [Go Module system][2], there's no need to use the `GOPATH`,
the source code can be stored in any path.

### Building

To build the application, run:

    $ go build ./cmd/adapter

Go will download the required dependencies to compile the source code and generate
a binary called `adapter` in your current working directory.

_TIP_: You may prefer to use `go install` rather than `go build` to cache build
artifacts and reduce future compile times. In this case the binary is placed in
`$GOPATH/bin/adapter`.

### Running the unit tests

Once you have the application building, you can run all the unit tests for the
project:

    $ go test ./...

To run the tests for a single package, change to package directory and run:

    $ go test .

_TIP_: It is recommended to run the tests with the [Go Race Detector][3] enabled.
You can enable it with `-race`, for example `go test -race ./...` will run the tests
in the whole project and will check for race conditions.

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

You can read more about the available images in the [tagging][5] document.

### Build an image

To build an image of your change using the applications's `Dockerfile`, run
these commands (replacing the repository host and tag with your own):

    $ docker build -t docker.io/yourusername/k8s-sysdig-adapter:latest .
    $ docker push docker.io/yourusername/k8s-sysdig-adapter:latest

### Verify your change

Deploy the image you've built to verify the change. More instructions on this
regards will be added soon.

[1]: https://golang.org/dl/
[2]: https://blog.golang.org/using-go-modules
[3]: https://blog.golang.org/race-detector
[4]: https://github.com/draios/kubernetes-sysdig-metrics-apiserver/issues
[5]: docs/tagging.md
