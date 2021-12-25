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

func main() {
	dir := flag.String("dir", ".", "Optional.  Searches every file in the directory for a match.  Non-recursive.")
	filename := flag.String("filename", "", "Optional.  Takes precedence over directory searches.")
	filetype := flag.String("filetype", ".md", "Only searches files of this type.")
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
