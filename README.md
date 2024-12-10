# CCC:=curl-command-cleaner

ccc is a tool designed to allow you to tweak curl commands (remove unnecessary headers and cookies....).

It features:

* full parsing logic (from https://blog.gopheracademy.com/advent-2014/parsers-lexers/)
* "shaking mode" := curl command is run once, output is saved and compared byte to byte to output of a similar curl command with (at least one) flag removed (based on some heuristics)?
* storing commands in a sqlite database
* a web UI

# Usage

```bash
    go run main.go https://jsonplaceholder.typicode.com/photos
```


