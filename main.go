package main

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"net/http"
	"os"
	"os/exec"
	regexp "regexp"
)

const (
	VCSGit        = "git"
	VCSBazaar     = "bzr"
	VCSFossil     = "fossil"
	VCSMercurial  = "hg"
	VCSSubversion = "svn"
	regex         = `([a-zA-Z0-9\.\/\-]*)\s*(git|bzr|fossil|svn|hg)\s(.*)`
)

var GoImportRegex *regexp.Regexp

var helpText = `Usage: ggs {package} [path]

ggs, or "go get source" is a tool used to download the source code 
of a Go module. 

ggs will clone the default branch from the specified repository, e.g. "main" or "master".

ggs places downloaded source code, by default, into $GOPATH/src/[package].
Optionally, you can define a filesystem path, and ggs will
instead put the code into that location.
`

var outPath = ""

func init() {
	GoImportRegex = regexp.MustCompile(regex)
}

func main() {
	// invoked as ggs {package} [proxy]
	// e.g ggs k8s.io/client-go

	if len(os.Args) < 2 {
		fmt.Println(helpText)
		os.Exit(1)
	}

	if len(os.Args) > 2 {
		outPath = os.Args[2]
	} else {
		outPath = fmt.Sprintf("%s/%s/%s", os.Getenv("GOPATH"), "src", os.Args[1])
	}

	var httpErr, httpsErr error
	var res *http.Response
	// try https first
	res, httpsErr = http.Get(fmt.Sprintf("https://%s?go-get=1", os.Args[1]))
	if httpsErr != nil {
		res, httpErr = http.Get(fmt.Sprintf("http://%s?go-get=1", os.Args[1]))
		if httpErr != nil {
			fmt.Printf("error getting HTTP response from %s, tried HTTP(S).\n%s", os.Args[1], errors.Join(httpsErr, httpErr).Error())
			os.Exit(1)
		}
	}

	tokenizer := html.NewTokenizer(res.Body)
	var found = false
	var name, content string
	for {
		if found {
			break
		}
		switch tokenizer.Next() {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				fmt.Printf("unable to parse module information from url %s\n", os.Args[1])
				os.Exit(1)
			} else {
				fmt.Printf("error parsing HTML response body: %s", tokenizer.Err())
				os.Exit(1)
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			tok := tokenizer.Token()
			if tok.DataAtom == atom.Meta {
				// get name value
				name, content = parseNameAndContent(tok)
				if name == "go-import" {
					found = true
					break
				}
			}
		}
	}

	matches := GoImportRegex.FindStringSubmatch(content)
	if len(matches) != 4 {
		fmt.Printf("invalid go-import formatting, could not extract package, vcs, and url. content was: %s", content)
		os.Exit(1)
	}

	var pkg, vcs, url = matches[1], matches[2], matches[3]
	if pkg != os.Args[1] {
		fmt.Printf("return pkg value %s does not match input %s\n", pkg, os.Args[1])
		os.Exit(1)
	}

	var command = "clone"
	switch vcs {
	case VCSSubversion:
		command = "checkout"
	case VCSBazaar:
		command = "branch" // maybe? nobody uses bazaar so idk
	}

	cmd := exec.Command(vcs, command, url, outPath)
	errPipe, _ := cmd.StderrPipe()
	outPipe, _ := cmd.StdoutPipe()
	go io.Copy(os.Stderr, errPipe)
	go io.Copy(os.Stdout, outPipe)

	cmd.Start()
	cmd.Wait()
}

func parseNameAndContent(tok html.Token) (string, string) {
	var name, content string
	for _, a := range tok.Attr {
		if a.Key == "name" {
			name = a.Val
		}
		if a.Key == "content" {
			content = a.Val
		}
	}

	return name, content
}
