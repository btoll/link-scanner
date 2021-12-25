package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

type Link struct {
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
}

/*
func getHead(c chan Link, link string) {
	_, err := url.ParseRequestURI(link)
	if err != nil {
		c <- Link{
			URL:        link,
			StatusCode: -1,
		}
		return
	}

	resp, err := http.Head(link)
	if err != nil {
		c <- Link{
			URL:        link,
			StatusCode: 401,
		}
		return
	}

	c <- Link{
		URL:        link,
		StatusCode: resp.StatusCode,
	}
}
*/

func getHead(link string) Link {
	_, err := url.ParseRequestURI(link)
	if err != nil {
		return Link{
			URL:        link,
			StatusCode: -1,
		}
	}

	resp, err := http.Head(link)
	if err != nil {
		return Link{
			URL:        link,
			StatusCode: 401,
		}
	}

	return Link{
		URL:        link,
		StatusCode: resp.StatusCode,
	}
}

func (ls *LinkScanner) getLinkScannerSession() (ScannerSession, error) {
	var files []string
	var links []Link
	var err error
	var numFiles int

	urls := make(map[string][]Link, 1000)

	//	c := make(chan Link, len(urls))

	if ls.Dir != "" {
		files, err = ls.readDirByFileType()
		if err != nil {
			return ScannerSession{}, err
		}

		for _, filename := range files {
			files, err := ls.readFileByPattern(filename)
			if err != nil {
				fmt.Println(err)
			}

			//			urls[filename] = fs
			for _, file := range files {
				links = append(links, getHead(file))
			}

			urls[filename] = links
			numFiles += 1
		}
	}

	/*
		for _, v := range urls {
			for _, url := range v {
				go getHead(c, url)
			}
		}
	*/

	/*
		for _, filename := range files {

			//		fmt.Println("filename", filename)
			//			links = append(links, <-c)
			for i := 0; i < len(v); i++ {
				links = append(links, <-c)
			}

			urls[key] = links
		}
	*/

	return ScannerSession{
		TotalFiles: numFiles,
		Tree:       urls,
		Links:      links,
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

	for filename, links := range ss.Tree {
		fmt.Printf("\n%s\n", filename)

		for _, link := range links {
			fmt.Printf("\t%s\n", link)
		}
		//		if link.StatusCode == 401 {
		//			fmt.Println(link)
		//		}
	}
}
