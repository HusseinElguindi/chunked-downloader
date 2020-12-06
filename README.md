# Chunked Downloader
Chunked downloader that speeds up file downloads by using byte ranges.

![Downloader Demo Image](https://github.com/HusseinElguindi/chunked-downloader/blob/master/img/demo.png?raw=true)

## Features
- [x] Multithreading (goroutines)
- [x] Command-line flags
- [x] Progress bar per thread
- [x] Editable read/write size
- [x] Automatic filename from URL
- [ ] Pause and resume downloads
- [ ] Download more than 1 file in the same instance

## Getting Started
To start, clone the repo. Now, you can run the program directly or build it into an executable.
>- **Option 1:** Running the program directly
```shell
$ go run .
```
>- **Option 2:** Building the program
```shell
$ go build .
$ chunked-downloader.exe (Windows)
$ ./chunked-downloader (Linux and MacOS)
```
Once the program is running, everything is self-explanatory and intuitive.

![Downloader Help Command](https://github.com/HusseinElguindi/chunked-downloader/blob/master/img/help.png?raw=true)

## Author
- **Hussein Elguindi** - *all the work*

## License 
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
- [GPL-3.0 License](https://www.gnu.org/licenses/gpl-3.0)
- Copyright 2020 © Hussein Elguindi

