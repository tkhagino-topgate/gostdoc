package main

// from stringer. https://godoc.org/golang.org/x/tools/cmd/stringer

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/favclip/genbase"
	"github.com/tkhagino-topgate/gostdoc"
)

var (
	typeNames          = flag.String("type", "", "comma-separated list of type names; must be set")
	output             = flag.String("output", "", "output file name; default stdout")
	format             = flag.String("format", "tsv", "output format; tsv, tsvshort, json")
	ignoreStructSuffix = flag.String("ignore-struct-suffix", "", "comma-separated list; ignore struct suffix")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tgostdoc [flags] [directory]\n")
	fmt.Fprintf(os.Stderr, "\tgostdoc [flags] files... # Must be a single package\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("gostdoc: ")
	flag.Usage = Usage
	flag.Parse()

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	outputFormat := gostdoc.OutputFormatType(*format)
	if outputFormat != gostdoc.OutputFormatTypeTsv && outputFormat != gostdoc.OutputFormatTypeTsvShort && outputFormat != gostdoc.OutputFormatTypeJSON {
		Usage()
	}

	// Parse the package once.
	var dir string
	var pInfo *genbase.PackageInfo
	var err error
	p := &genbase.Parser{SkipSemanticsCheck: true}
	if len(args) == 1 && isDirectory(args[0]) {
		dir = args[0]
		pInfo, err = p.ParsePackageDir(dir)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		dir = filepath.Dir(args[0])
		pInfo, err = p.ParsePackageFiles(args)
		if err != nil {
			log.Fatal(err)
		}
	}

	// var typeInfos genbase.TypeInfos
	// if len(*typeNames) == 0 {
	// 	typeInfos = pInfo.CollectTaggedTypeInfos("+gostdoc")
	// } else {
	// 	typeInfos = pInfo.CollectTypeInfos(strings.Split(*typeNames, ","))
	// }
	var typeInfos = pInfo.TypeInfos()

	if len(typeInfos) == 0 {
		flag.Usage()
	}

	suffix := make([]string, 0, 10)
	for _, v := range strings.Split(*ignoreStructSuffix, ",") {
		s := strings.TrimSpace(v)
		if s != "" {
			suffix = append(suffix, s)
		}
	}

	bu, err := gostdoc.Parse(pInfo, typeInfos, &gostdoc.ParseOptions{
		IgnoreStructSuffix: suffix,
	})
	if err != nil {
		log.Fatal(err)
	}

	outputOpt := gostdoc.OutputOptions{
		Format: outputFormat,
	}

	// Format the output.
	src, err := bu.Emit(&outputOpt)
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
	}

	// Write to file.
	outputName := *output
	if outputName == "" {
		_, err = os.Stdout.Write(src)
		if err != nil {
			log.Fatalf("writing stdout: %s", err)
		}
	} else {
		err = ioutil.WriteFile(outputName, src, 0644)
		if err != nil {
			log.Fatalf("writing output: %s", err)
		}
	}
}

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}
