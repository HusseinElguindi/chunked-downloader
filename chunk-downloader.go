package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
	f, _ := os.OpenFile("test.mp4", os.O_CREATE|os.O_RDWR, 0666)
	defer f.Close()

	t := Task{
		URL:    "https://cdn.discordapp.com/attachments/574963057957011476/615978766442692613/profileconvert.mp4",
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
	// chunkRemainder := contentSize % int64(t.chunks)

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
}

// download bytes from Task struct's URL, with ranges (for chunks)
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

// write downloaded bytes to Task struct's outfile, with offset for ranged downloads
func (t *Task) writeBytes(start int64) {

}

// perform validation tests to file, to compare download with origninal file
func (t *Task) validateDownload() {

}

func handleErr(err *error) {
	if *err != nil {
		panic(err)
	}
}
