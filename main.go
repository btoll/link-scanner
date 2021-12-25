package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type Link struct {
	Owner      string
	URL        string
	StatusCode int
}

type LinkScanner struct {
	Dir      string
	FileName string
	FileType string
}

type ScannerSession struct {
	TotalFiles int
	Tree       map[string][]Link
	Links      []Link
	Failed     []Link
}

func getHead(c chan Link, l Link) {
	_, err := url.ParseRequestURI(l.URL)
	if err != nil {
		l.StatusCode = -1
		c <- l
		return
	}

	resp, err := http.Head(l.URL)
	if err != nil {
		l.StatusCode = 401
		c <- l
		return
	}

	l.StatusCode = resp.StatusCode
	c <- l
}

func (ls *LinkScanner) getLinkScannerSession() (ScannerSession, error) {
	var files []string
	var links []Link
	var failed []Link
	var err error
	var numFiles int

	urls := make(map[string][]string, 1000)

	if ls.Dir != "" {
		files, err = ls.readDirByFileType()
		if err != nil {
			return ScannerSession{}, err
		}

		for _, filename := range files {
			fs, err := ls.readFileByPattern(filename)
			if err != nil {
				fmt.Println(err)
			}

			urls[filename] = fs

			for _, f := range fs {
				files = append(files, f)
			}

			numFiles += 1
		}
	}

	c := make(chan Link, len(urls))

	for filename, v := range urls {
		for _, url := range v {
			l := Link{
				Owner: filename,
				URL:   url,
			}
			go getHead(c, l)
		}
	}

	for _, v := range urls {
		for i := 0; i < len(v); i++ {
			link := <-c
			links = append(links, link)
			if link.StatusCode == 401 {
				failed = append(failed, link)
			}
		}
	}

	return ScannerSession{
		TotalFiles: numFiles,
		Links:      links,
		Failed:     failed,
	}, err
}

func (ls *LinkScanner) readDirByFileType() ([]string, error) {
	var files []string

	dirEntries, err := os.ReadDir(ls.Dir)
	if err != nil {
		return files, err
	}

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			name := dirEntry.Name()
			fileExtension := len(name) - len(ls.FileType)

			if name[fileExtension:] == ls.FileType {
				files = append(files, name)
			}
		}
	}

	return files, err
}

func (ls *LinkScanner) readFileByPattern(filename string) ([]string, error) {
	var filenames []string

	readFile, err := os.Open(fmt.Sprintf("%s/%s", ls.Dir, filename))
	if err != nil {
		return filenames, err
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	re := regexp.MustCompile(`(?:https?:\/\/.*\.[^\W)'"][\w\-/?=]*)`)
	for fileScanner.Scan() {
		text := fileScanner.Text()
		if re.MatchString(text) {
			submatchall := re.FindAllString(text, -1)
			for _, element := range submatchall {
				filenames = append(filenames, element)
			}
		}
	}

	readFile.Close()

	return filenames, err
}

func main() {
	dir := flag.String("dir", "", "If given, searches every file for a match.")
	filename := flag.String("filename", "", "If given a directory, will only search by this filetype.  Defaults to `md`.")
	filetype := flag.String("filetype", ".md", "If given a directory, will only search by this filetype.  Defaults to `md`.")
	flag.Parse()

	urlIgnore := `\.onion|example\.com|<YOUR_DOMAIN>`

	ls := LinkScanner{
		Dir:      *dir,
		FileName: *filename,
		FileType: *filetype,
	}

	ss, err := ls.getLinkScannerSession()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(ss.Failed) > 0 {
		split := strings.Split(urlIgnore, "|")
		fmt.Printf("[FAILED] Number of failures = %d\n", len(ss.Failed))
		fmt.Println("Ignored all URLs that contained any of the following:")
		fmt.Println(strings.Join(split, "\n"))
		re := regexp.MustCompile(urlIgnore)
		for _, link := range ss.Failed {
			if re.MatchString(link.URL) {
				continue
			}
			if link.StatusCode == 401 {
				fmt.Printf("\n(%d) %s\tOwner: %s", link.StatusCode, link.URL, link.Owner)
			}
		}
		// Add newline so it's not on the same line as the cursor.
		fmt.Println()
		os.Exit(1)
	} else {
		fmt.Println("No failures.")
	}
}
