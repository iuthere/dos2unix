package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/iuthere/dos2unix/scan"
	"github.com/mattn/go-zglob"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
)

type config struct {
	recursive bool
	write     bool
	verbose   bool
}

func skipDir(name string) bool {
	switch name {
	case ".git", ".idea", ".vscode":
		return true
	case "node_modules":
		return true
	}
	return false
}

func main() {
	cfg := readConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && path != "." {
			if !cfg.recursive {
				return fs.SkipDir
			}
			if skipDir(d.Name()) {
				return fs.SkipDir
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if !d.IsDir() && d.Type().IsRegular() {
			process := false
			for _, s := range flag.Args() {
				if b, _ := zglob.Match(s, path); b {
					process = true
					break
				}
				if b, _ := zglob.Match(s, d.Name()); b {
					process = true
					break
				}

			}
			if process {
				err := processFile(ctx, path, cfg)
				if err != nil {
					fmt.Println("-", err)
				}
			}
		}
		return nil
	})
}

// readConfig attempts to read flags or exists
// upon error.
func readConfig() *config {
	r := flag.Bool("r", false, "visit folders recursively.")
	h := flag.Bool("h", false, "print help.")
	w := flag.Bool("w", false, "actually write \\r\\n to \\n changes.")
	v := flag.Bool("v", false, "verbose about wrong file types.")
	flag.Parse()
	if *h || len(flag.Args()) == 0 {
		name := filepath.Base(os.Args[0])
		name = strings.TrimSuffix(name, filepath.Ext(name))

		fmt.Fprintf(flag.CommandLine.Output(), "dos2unix converts line endings from \\r\\n to \\n in text files.\nUsage:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "%[1]s <filePattern(s)>       report only in the current folder.\n", name)
		fmt.Fprintf(flag.CommandLine.Output(), "%[1]s -r <filePattern(s)>    report only in the current folder and recursively.\n", name)
		fmt.Fprintf(flag.CommandLine.Output(), "%[1]s -w -r <filePattern(s)> replace in the current folder and recursively.\n", name)
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "  <filePattern(s)>    file pattern or space-separated list of file patterns, ex. *.tmpl\n")
		os.Exit(0)
	}
	for _, s := range flag.Args() {
		_, err := filepath.Match(s, ".")
		errs := false
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "Parameter error: invalid pattern: %v\n", s)
			errs = true
		}
		if errs {
			os.Exit(1)
		}
	}
	return &config{
		recursive: *r,
		write:     *w,
		verbose:   *v,
	}
}

// processFile converts the file if cfg.write == true,
// otherwise simply reports.
func processFile(ctx context.Context, path string, cfg *config) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening %v: %v", path, err)
	}
	defer file.Close()

	var temp *os.File
	if cfg.write {
		temp, err = ioutil.TempFile("", "*.tmp")
		if err != nil {
			return fmt.Errorf("unable to create a tmp file: %v", err)
		}
	}
	if temp != nil {
		defer func() {
			//temp.Sync()
			temp.Close()
			_ = os.Remove(temp.Name())
		}()
	}

	firstBlock := make([]byte, 512)
	read, err := file.Read(firstBlock)
	if err != nil && err != io.EOF {
		return fmt.Errorf("can't read first 512 bytes: %v", err)
	}
	typeDetected := http.DetectContentType(firstBlock[:read])
	if _, err = file.Seek(0, 0); err != nil {
		return fmt.Errorf("can't seek the file: %v", err)
	}

	if strings.HasPrefix(typeDetected, "text/") {
		scanner := bufio.NewScanner(file)
		scanner.Split(scan.ScanLinesKeep)
		found := false
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			line := scanner.Text()
			if strings.HasSuffix(line, "\r\n") {
				found = true
				line = strings.Replace(line, "\r\n", "\n", 1)
			}
			if temp == nil && found {
				break // As we only perform scanning, we already know the answer
			}
			if temp != nil {
				_, err := temp.WriteString(strings.Replace(line, "\r\n", "\n", 1))
				if err != nil {
					return fmt.Errorf("error writing to temp file %v", err)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading %v: %v", path, err)
		} else {
			if found {
				if cfg.write {
					err = file.Close()
					if err != nil {
						return fmt.Errorf("can't close original file: %v", err)
					}
					destination, err := os.Create(path)
					if err != nil {
						return fmt.Errorf("error opening %v: %v", path, err)
					}
					defer func() {
						destination.Sync()
						destination.Close()
					}()

					if err = temp.Sync(); err != nil {
						return fmt.Errorf("can't sync to temp file: %v", err)
					}
					if err = temp.Close(); err != nil {
						return fmt.Errorf("can't close temp file: %v", err)
					}
					source, err := os.Open(temp.Name())
					if err != nil {
						return fmt.Errorf("error creating fresh file: %v", err)
					}
					defer source.Close()
					writer := bufio.NewWriter(destination)
					if _, err = io.Copy(writer, source); err != nil {
						return fmt.Errorf("error writing to destination file: %v", err)
					}
					if err = writer.Flush(); err != nil {
						return fmt.Errorf("error flushing to destination file: %v", err)
					}
					fmt.Printf("+ removed \\r\\n:      %v\n", path)
				} else {
					fmt.Printf("+ contains \\r\\n:     %v\n", path)
				}
			}
		}
	} else {
		if cfg.verbose {
			return fmt.Errorf("error reading %v: wrong file type %v", path, typeDetected)
		}
	}
	return nil
}
