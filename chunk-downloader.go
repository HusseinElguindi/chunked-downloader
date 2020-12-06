// Hussein Elguindi

package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

const (
	// ReadSize - Number of bytes to read/write from download buffer (per thread)
	ReadSize = int64(8 * 1024 * 1024) // 1KiB
)

var (
	err *error
)

// Task struct for chunk downloads
type Task struct {
	URL      string
	chunks   int
	File     *os.File
	wg       *sync.WaitGroup
	bar      *mpb.Bar
	readSize int64
}

// Main should act as another file, calling this download command
// func main() {
// 	f, _ := os.OpenFile("test.mp4", os.O_CREATE|os.O_RDWR, 0666)
// 	defer f.Close()

// 	t := Task{
// 		URL:    "http://distribution.bbb3d.renderfarming.net/video/mp4/bbb_sunflower_1080p_60fps_stereo_abl.mp4",
// 		chunks: 10,
// 		File:   f,
// 		wg:     nil,
// 	}

// 	t.initDownload()
// }

func (t *Task) initDownload() {
	// get headers to initiate chunked download or normal download
	r, err := http.Head(t.URL)
	handleErr(&err)

	contentSizeString := r.Header.Get("Content-Length")
	if len(contentSizeString) < 1 {
		fmt.Println("no content length provided: continuing download without multithreading")
		// return

		t.chunks = 1
		contentSizeString = "0"
	}

	// if r.Header.Get("Accept-Ranges") != "bytes" {
	if !strings.Contains(r.Header.Get("Accept-Ranges"), "bytes") {
		fmt.Println("accept range not bytes: continuing download without multithreading")
		// return

		t.chunks = 1
	}

	contentSize, err := strconv.ParseInt(contentSizeString, 10, 64)
	handleErr(&err)

	// p := mpb.New(mpb.WithWidth(64))
	p := mpb.New()

	chunkSize := contentSize / int64(t.chunks)

	var start, end int64

	t.wg = &sync.WaitGroup{}
	t.wg.Add(t.chunks)

	var bar *mpb.Bar
	var i int
	for i = 1; i < t.chunks; i++ {
		bar = p.AddBar(chunkSize,
			mpb.PrependDecorators(decor.Name(fmt.Sprintf("Thread %02d: ", i)), decor.Counters(decor.UnitKiB, "%.1f / %.1f", decor.WC{W: 20, C: decor.DidentRight})),
			mpb.AppendDecorators(decor.Percentage(decor.WC{W: 10, C: decor.DidentRight})),
		)

		end += chunkSize
		go t.downloadBytes(start, end-1, chunkSize, bar)
		start = end
	}

	if i > 1 {
		bar = p.AddBar(chunkSize,
			mpb.PrependDecorators(decor.Name(fmt.Sprintf("Thread %02d: ", i)), decor.Counters(decor.UnitKiB, "%.1f / %.1f", decor.WC{W: 20, C: decor.DidentRight})),
			mpb.AppendDecorators(decor.Percentage(decor.WC{W: 10, C: decor.DidentRight})),
		)
	}
	chunkSize = contentSize - start
	end += chunkSize
	go t.downloadBytes(start, end, chunkSize, bar)

	t.bar = p.AddBar(contentSize,
		mpb.PrependDecorators(decor.Name("TOTAL:     "), decor.Counters(decor.UnitKiB, "%.1f / %.1f", decor.WC{W: 20, C: decor.DidentRight})),
		mpb.AppendDecorators(decor.Percentage(decor.WC{W: 10, C: decor.DidentRight})),
	)

	t.wg.Wait()
	if t.bar != nil {
		t.bar.SetCurrent(contentSize)
	}
	p.Wait()

	t.validateDownload(contentSize)
}

// download bytes from Task struct's URL and write to file, with ranges (for chunks)
func (t *Task) downloadBytes(start int64, end int64, chunkSize int64, bar *mpb.Bar) {
	// fmt.Printf("range: %d - %d\n", start, end)

	req, err := http.NewRequest("GET", t.URL, nil)
	handleErr(&err)
	if end > 0 {
		req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	}

	client := new(http.Client)
	res, err := client.Do(req)
	handleErr(&err)

	defer res.Body.Close()
	defer t.wg.Done()

	buf := make([]byte, t.readSize)
	var wrote int64
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

		wrote = int64(nw)
		start += wrote

		if bar != nil {
			bar.IncrInt64(wrote)
		}
		if t.bar != nil {
			t.bar.IncrInt64(wrote)
		}
	}

	if bar != nil {
		bar.SetCurrent(chunkSize)
	}
}

// perform validation test, to compare download with origninal file
func (t *Task) validateDownload(contentSize int64) {
	if contentSize < 1 {
		fmt.Println("Download completed.")
		return
	}

	i, err := t.File.Stat()
	handleErr(&err)

	if i.Size() == contentSize {
		fmt.Println("Filesizes match: download completed.")
	} else {
		fmt.Println("Download completed.")
		// 	fmt.Println("Filesizes did not match: download may be altered.")
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
		panic(*err)
	}
}
