package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

/* the program */
type LinkScanner struct {
	Dir       string
	FileName  string
	FileType  string
	LinkRegex string
	//	SkipCode    string
	SkipPattern string
	Header      http.Header
}

type Link struct {
	Owner      string
	URL        string
	StatusCode int
	Error      error
}

/* the results of the program */
type ScannerSession struct {
	TotalFiles int
	Tree       map[string][]Link
	Links      []Link
	Failed     []Link
	Skipped    []Link
}

func (ls *LinkScanner) getLinkScannerSession() (ScannerSession, error) {
	var files []string
	var links []Link
	var failed []Link
	var skipped []Link
	var err error
	var numFiles int

	urls := make(map[string][]string, 1000)

	if ls.FileName != "" {
		files, err = ls.readFileByPattern(ls.FileName)
		if err != nil {
			return ScannerSession{}, err
		}
		urls[ls.FileName] = files
		numFiles = 1
	} else if ls.Dir != "" {
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
			go getHead(c, l, ls.Header)
		}
	}

	re := regexp.MustCompile(ls.SkipPattern)
	for _, v := range urls {
		for i := 0; i < len(v); i++ {
			link := <-c
			links = append(links, link)
			if !(link.StatusCode >= 200 && link.StatusCode < 400) {
				if re.MatchString(link.URL) {
					skipped = append(skipped, link)
				} else {
					failed = append(failed, link)
				}
			}
		}
	}

	return ScannerSession{
		TotalFiles: numFiles,
		Links:      links,
		Failed:     failed,
		Skipped:    skipped,
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
			filename := dirEntry.Name()
			fileExtension := len(filename) - len(ls.FileType)

			if filename[fileExtension:] == ls.FileType {
				files = append(files, fmt.Sprintf("%s/%s", strings.TrimSuffix(ls.Dir, "/"), filename))
			}
		}
	}

	return files, err
}

func (ls *LinkScanner) readFileByPattern(filename string) ([]string, error) {
	var filenames []string

	readFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	re := regexp.MustCompile(ls.LinkRegex)
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
