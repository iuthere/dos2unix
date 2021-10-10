[![https://github.com/iuthere/dos2unix](./doc/gobadge.svg)](https://pkg.go.dev/github.com/iuthere/dos2unix)

> **Warning**: It's still an experimental tool far from the production ready use. Please use with care. Create backup or use only in VCS-enabled folders. Only tested in Windows.

# dos2unix

Converts text files with \r\n line endings into \n. Gracefully interrupts upon Ctrl+C. Recognition whether a file consists of a text or not is performed using [http.DetectContentType](https://golang.org/pkg/net/http/#DetectContentType) based on the algorithm described at https://mimesniff.spec.whatwg.org/.

Additions and bug reports are welcome!

Requirement: `Go 1.16`.

## Install

```
> go install github.com/iuthere/dos2unix@latest
```

## Use

```
> dos2unix
dos2unix converts line endings from \r\n to \n in text files.
Usage:
dos2unix <filePattern(s)>       report only in the current folder.
dos2unix -r <filePattern(s)>    report only in the current folder and recursively.
dos2unix -w -r <filePattern(s)> replace in the current folder and recursively.
Flags:
  -h    print help.
  -r    visit folders recursively.
  -v    verbose about wrong file types.
  -w    actually write \r\n to \n changes.
  <filePattern(s)>    file pattern or space-separated list of file patterns, ex. *.tmpl
```

## Example

Recursively show all supposedly text files containing \r\n line endings.

```
> dos2unix -r *
```

Apply conversion to all supposedly text files containing \r\n line endings.

```
> dos2unix -r -w *
```

## Todo

* Recognize actual paths vs patterns.
* Rewrite scanner to support long lines (bufio.Scanner: token too long).
