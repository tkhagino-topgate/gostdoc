# gostdoc
gostdoc

thanks to [jwg](https://github.com/favclip/jwg)

## Description
analysis and documentation tool for golang's structure

### Example

## Installation
```
$ go get github.com/tkhagino-topgate/gostdoc/cmd/gostdoc
$ gostdoc
Usage of gostdoc:
	gostdoc [flags] [directory]
	gostdoc [flags] files... # Must be a single package
Flags:
  -format string
    	output format; tsv, tsvshort, json (default "tsv")
  -ignore-struct-suffix string
    	comma-separated list; ignore struct suffix
  -output string
    	output file name; default stdout
  -type string
    	comma-separated list of type names; must be set
```

## Command sample
