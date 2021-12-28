package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func getHead(c chan Link, l Link) {
	_, err := url.ParseRequestURI(l.URL)
	if err != nil {
		l.StatusCode = -1
		l.Error = err
		c <- l
		return
	}

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

	urlIgnore := `\.onion|example\.com`

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

	fmt.Printf("[INFO] Number of files scanned    = %d\n", ss.TotalFiles)
	fmt.Printf("[INFO] Number of pattern matches  = %d\n", len(ss.Links))
	fmt.Printf("[INFO] Number of pattern failures = %d\n", len(ss.Failed))
	fmt.Printf("[INFO] Ignore pattern list:\n%s\n", strings.Join(strings.Split(urlIgnore, "|"), "\n"))

	ignored := 0
	if len(ss.Failed) > 0 {
		re := regexp.MustCompile(urlIgnore)
		for _, link := range ss.Failed {
			if re.MatchString(link.URL) {
				ignored += 1
				continue
			}
			if link.StatusCode == 401 {
				fmt.Printf("\n(%d)   (Link)  %s\n\t(Error) %s\n\t(Owner) %s\n", link.StatusCode, link.URL, link.Error, link.Owner)
			}
		}

		fmt.Printf("[INFO] Number ignored = %d\n", ignored)

		if ignored != len(ss.Failed) {
			os.Exit(1)
		}
	} else {
		fmt.Println("[SUCCESS] No failures.")
	}
}
