package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func getHead(c chan Link, l Link) {
	/*
		_, err := url.ParseRequestURI(l.URL)
		if err != nil {
			l.StatusCode = -1
			l.Error = err
			c <- l
			return
		}
	*/

	resp, err := http.Head(l.URL)
	if err != nil {
		l.StatusCode = 401
		l.Error = err
		c <- l
		return
	}

	l.StatusCode = resp.StatusCode
	c <- l
}

func main() {
	dir := flag.String("dir", "", "Optional.  Searches every file in the directory for a match.  Non-recursive.")
	filename := flag.String("filename", "", "Optional.  Takes precedence over directory searches.")
	filetype := flag.String("filetype", ".md", "Only searches files of this type.  Include the period, i.e., `.html`")
	flag.Parse()

	ls := LinkScanner{
		Dir:         *dir,
		FileName:    *filename,
		FileType:    *filetype,
		SkipPattern: `\.onion|example\.com`,
	}

	ss, err := ls.getLinkScannerSession()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("[INFO] Number of files scanned = %d\n", ss.TotalFiles)
	fmt.Printf("[INFO] Number of links matched = %d\n", len(ss.Links))
	fmt.Printf("[INFO] Links failed  = %d\n", len(ss.Failed))
	fmt.Printf("[INFO] Links skipped = %d\n", len(ss.Skipped))
	fmt.Printf("[INFO] Skip pattern list:\n%s\n", strings.Join(strings.Split(ls.SkipPattern, "|"), "\n"))

	/*
		if len(ss.Skipped) > 0 {
			for _, link := range ss.Skipped {
				fmt.Printf("\n(%d)   (Link)  %s\n\t(Error) %s\n\t(Owner) %s\n", link.StatusCode, link.URL, link.Error, link.Owner)
			}
		}
	*/

	if len(ss.Failed) > 0 {
		for _, link := range ss.Failed {
			fmt.Printf("\n(%d)   (Link)  %s\n\t(Error) %s\n\t(Owner) %s\n", link.StatusCode, link.URL, link.Error, link.Owner)
		}

		if len(ss.Skipped) != len(ss.Failed) {
			os.Exit(1)
		}
	} else {
		fmt.Println("[SUCCESS] No failures.")
	}
}
