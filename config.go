package main

import (
	"flag"
	"fmt"
	"strings"
)

type Config struct {
	entryPath  string
	extensions []string
	ignore     []string
}

func loadConfig() *Config {
	path := flag.String("path", ".", "Path to directory that includes file samples. Defaults to current directory.")
	ext := flag.String("ext", "*", "Comma separated file extensions that will be parsed. Defaults to all extensions.")
	ignore := flag.String("ignore", "", "Comma separated strings. Paths which contain these will be ignored.")
	flag.Parse()

	var extensions []string
	if *ext != "*" {
		extensions = strings.Split(*ext, ",")
		for i, e := range extensions {
			extensions[i] = fmt.Sprintf(".%s", e)
		}
	}

	c := &Config{
		entryPath:  *path,
		extensions: extensions,
		ignore:     strings.Split(*ignore, ","),
	}
	return c
}
