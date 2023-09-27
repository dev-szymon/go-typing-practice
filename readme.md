## CLI tool for typing practice

This tool has been created to practice typing speed on texts generated from real code files instead of artificial words.

Build:
```
go build -o ./cmd
```

Use `-h` flag to check available flags:
```
./cmd/go-typing-practice -h
```
Output:
```
Usage of ./cmd/go-typing-practice:
  -ext string
        Comma separated file extensions that will be parsed. Defaults to all extensions. (default "*")
  -ignore string
        Comma separated strings. Paths which contain these will be ignored.
  -path string
        Path to directory that includes file samples. Defaults to current directory. (default ".")
```

Example:
```
./cmd/go-typing-practice -path=../samples -ext=go,ts,tsx -ignore=node_modules
```