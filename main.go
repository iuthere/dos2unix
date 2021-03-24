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

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	r := flag.Bool("r", false, "visit folders recursively")
	h := flag.Bool("h", false, "print help")
	w := flag.Bool("w", false, "actually write")
	flag.Parse()
	if *h || len(flag.Args()) == 0 {
		name := filepath.Base(os.Args[0])
		name = strings.TrimSuffix(name, filepath.Ext(name))

		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "%[1]s <filePattern(s)>       report only in the current folder.\n", name)
		fmt.Fprintf(flag.CommandLine.Output(), "%[1]s -r <filePattern(s)>    report only in the current folder and recursivly.\n", name)
		fmt.Fprintf(flag.CommandLine.Output(), "%[1]s -w -r <filePattern(s)> replace in the current folder and recursivly.\n", name)
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "  <filePattern(s)>    file pattern or space-separated list of file patterns, ex. *.tmpl\n")
		os.Exit(0)
	}

	for _, s := range flag.Args() {
		_, err := filepath.Match(s, ".")
		errs := false
		if err != nil {
			fmt.Printf("invalid pattern: %v\n", s)
			errs = true
		}
		if errs {
			os.Exit(1)
		}
	}

	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && path != "." {
			if !*r {
				return fs.SkipDir
			}
			switch d.Name() {
			case ".git":
				return fs.SkipDir
			case "node_modules":
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
				err := processFile(ctx, path, d, *w)
				if err != nil {
					fmt.Println("-", err)
				}
			}
		}
		return nil
	})
}

func processFile(ctx context.Context, path string, d fs.DirEntry, write bool) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening %v: %v\n", path, err)
	}
	defer file.Close()

	temp, err := ioutil.TempFile("", "*.tmp")
	if err != nil {
		return fmt.Errorf("unable to create a tmp file: %v", err)
	}
	defer func() {
		//temp.Sync()
		temp.Close()
		_ = os.Remove(temp.Name())
	}()

	firstBlock := make([]byte, 512)
	read, err := file.Read(firstBlock)
	if err != nil && err != io.EOF {
		return fmt.Errorf("can't read first 512 bytes: %v", err)
	}
	typeDetected := http.DetectContentType(firstBlock[:read])
	if _, err = file.Seek(0, 0); err != nil {
		return fmt.Errorf("can't seek the file: %v", err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(scan.ScanLinesKeep)

	diff := false
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := scanner.Text()
		if strings.HasSuffix(line, "\r\n") {
			diff = true
			line = strings.Replace(line, "\r\n", "\n", 1)
		}
		_, err := temp.WriteString(strings.Replace(line, "\r\n", "\n", 1))
		if err != nil {
			return fmt.Errorf("error writing to temp file %v\n", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading %v: %v\n", path, err)
	} else {
		if strings.HasPrefix(typeDetected, "text/") {
			if diff {
				if write {
					err = file.Close()
					if err != nil {
						return fmt.Errorf("can't close original file: %v\n", err)
					}
					destination, err := os.Create(path)
					if err != nil {
						return fmt.Errorf("error opening %v: %v\n", path, err)
					}
					defer func() {
						destination.Sync()
						destination.Close()
					}()

					if err = temp.Sync(); err != nil {
						return fmt.Errorf("can't sync to temp file: %v\n", err)
					}
					if err = temp.Close(); err != nil {
						return fmt.Errorf("can't close temp file: %v\n", err)
					}

					source, err := os.Open(temp.Name())
					if err != nil {
						return fmt.Errorf("error creating fresh file: %v\n", err)
					}
					defer source.Close()

					writer := bufio.NewWriter(destination)

					if _, err = io.Copy(writer, source); err != nil {
						return fmt.Errorf("error writing to destination file: %v\n", err)
					}
					if err = writer.Flush(); err != nil {
						return fmt.Errorf("error flushing to destination file: %v\n", err)
					}
					fmt.Printf("+ removed \\r\\n:      %v\n", path)
				} else {
					fmt.Printf("+ contains \\r\\n:     %v\n", path)
				}
			} else {
				//fmt.Printf("nothing to change: %v\n", path)
			}
		} else {
			fmt.Printf("- wrong file type:   %v (%v)\n", path, typeDetected)
		}
	}
	return nil
}
