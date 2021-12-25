package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

type LinkScanner struct {
	Dir      string
	FileName string
	FileType string
}

type Link struct {
	Owner      string
	URL        string
	StatusCode int
}

type ScannerSession struct {
	TotalFiles int
	Tree       map[string][]Link
	Links      []Link
	Failed     []Link
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
