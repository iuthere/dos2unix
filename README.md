# dos2unix

Converts text files with \r\n line endings into \n. Gracefully interrupts upon Ctrl+C.

**Warning**: It's an experimental tool far from the production ready use. Please use with care. Create backup or use only in VCS-enabled folders. Only tested in Windows.

Additions and bug reports are welcome!

Requirement: `Go 1.16`.

## Install

```
> go install github.com/iuthere/dos2unix@latest
```

## Use

```
> dos2unix
Usage:
dos2unix <filePattern(s)>       report only in the current folder.
dos2unix -r <filePattern(s)>    report only in the current folder and recursivly.
dos2unix -w -r <filePattern(s)> replace in the current folder and recursivly.
  -h    print help
  -r    visit folders recursively
  -w    actually write
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
