# Go Get Source

`ggs` is a tool to download the source of a Go module.

## Installation

`go install github.com/ebauman/ggs`

Or download a binary. 

## Help Text

```
Usage: ggs {package} [path]

ggs, or "go get source" is a tool used to download the source code 
of a Go module. 

ggs will clone the default branch from the specified repository, e.g. "main" or "master".

ggs places downloaded source code, by default, into $GOPATH/src/[package].
Optionally, you can define a filesystem path, and ggs will
instead put the code into that location.
```

## Contributing

Toss up an issue + PR and let's chat about it. 