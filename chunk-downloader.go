package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

var (
	err *error
)

// Task struct for chunk downloads
type Task struct {
	URL    string
	chunks int
	File   *os.File
	wg     *sync.WaitGroup
}

// Main should act as another file, calling this download command
func main() {
	f, _ := os.OpenFile("test.zip", os.O_CREATE|os.O_RDWR, 0666)
	defer f.Close()

	t := Task{
		URL:    "http://ipv4.download.thinkbroadband.com/1GB.zip",
		chunks: 100,
		File:   f,
		wg:     nil,
	}

	t.initDownload()
}

func (t *Task) initDownload() {
	// get headers to initiate chunked download or normal download
	r, err := http.Head(t.URL)
	handleErr(&err)

	contentSizeString := r.Header.Get("Content-Length")
	if len(contentSizeString) < 1 {
		fmt.Println("no content length provided")
		return
	}

	if r.Header.Get("Accept-Ranges") != "bytes" {
		fmt.Println("accept range not bytes")
		return
	}

	contentSize, err := strconv.ParseInt(contentSizeString, 10, 64)
	handleErr(&err)

	chunkSize := contentSize / int64(t.chunks)
	fmt.Println(contentSize)

	var start, end int64

	t.wg = &sync.WaitGroup{}
	t.wg.Add(t.chunks)

	for i := 1; i < t.chunks; i++ {
		end += chunkSize
		go t.downloadBytes(start, end-1)
		start = end
	}
	end += (contentSize - start)
	go t.downloadBytes(start, end)

	t.wg.Wait()

	t.validateDownload(contentSize)
}

// download bytes from Task struct's URL and write to file, with ranges (for chunks)
func (t *Task) downloadBytes(start int64, end int64) {
	fmt.Printf("range: %d - %d\n", start, end)

	req, err := http.NewRequest("GET", t.URL, nil)
	handleErr(&err)
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	client := new(http.Client)
	res, err := client.Do(req)
	handleErr(&err)

	defer res.Body.Close()
	defer t.wg.Done()

	readSize := 32 * 1024
	buf := make([]byte, readSize)
	for {
		nr, err := res.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				t.File.WriteAt(buf[:nr], start)
				break
			}
			fmt.Println(err)
			return
		}

		nw, err := t.File.WriteAt(buf[:nr], start)
		handleErr(&err)

		start += int64(nw)
	}
}

// perform validation test, to compare download with origninal file
func (t *Task) validateDownload(contentSize int64) {
	i, err := t.File.Stat()
	handleErr(&err)

	if i.Size() == contentSize {
		fmt.Println("Filesizes match, download completed.")
	}
}

// extract filename from URL
func getFilename(URL string) string {
	URLStruct, err := url.Parse(URL)
	handleErr(&err)

	return filepath.Base(URLStruct.Path)
}

func handleErr(err *error) {
	if *err != nil {
		panic(err)
	}
}
