// Hussein Elguindi

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func main() {
	tempURL := flag.String("URL", "", "The URL to download a file from.")
	tempFilename := flag.String("f", "", "The filename to save the download as.")
	threads := flag.Int("c", runtime.NumCPU(), "Amount of threads/chunks to split up the download, if the file accepts ranges.")
	readSize := flag.Int64("r", ReadSize, "Advanced: Number of bytes to read/write from download buffer per thread")
	flag.Parse()

	var URL string
	if *tempURL != "" {
		URL = *tempURL
	} else {
		fmt.Print("Enter Download URL: ")
		_, err := fmt.Scanln(&URL)
		handleErr(&err)
		fmt.Print("\n")
	}

	var filename string
	if *tempFilename != "" {
		if strings.Contains(*tempFilename, ".") {
			filename = *tempFilename
		} else {
			i := strings.Split(getFilename(URL), ".")
			filename = fmt.Sprintf("%v.%v", *tempFilename, i[1])
		}
	} else {
		filename = getFilename(URL)
		if filename == "." {
			filename = "download"
		}
	}

	fmt.Printf("Filename: %v \nChunks/Threads: %d\n\n", filename, *threads)

	f, _ := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	defer f.Close()

	t := Task{
		URL:      URL,
		chunks:   *threads,
		File:     f,
		wg:       nil,
		readSize: *readSize,
	}

	t.initDownload()
}
