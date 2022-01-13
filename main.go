package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var statusCodes = map[int]string{
	400: "400 Bad Request",
	401: "401 Unauthorized",
	403: "403 Forbidden",
	404: "404 Not Found",
	500: "500 Internal Server Error",
	501: "501 Not Implemented",
	502: "502 Bad Gateway",
}

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
		l.StatusCode = 400
		l.Error = err
		c <- l
		return
	}

	l.StatusCode = resp.StatusCode
	c <- l
}

func printLinks(links []Link) {
	if len(links) > 0 {
		for _, link := range links {
			fmt.Printf("[DEBUG] - %s\n", link.URL)
		}
	}
}

func main() {
	dir := flag.String("dir", "", "Optional.  Searches every file in the directory for a match.  Non-recursive.")
	filename := flag.String("filename", "", "Optional.  Takes precedence over directory searches.")
	filetype := flag.String("filetype", ".md", "Only searches files of this type.  Include the period, i.e., `.html`")
	verbose := flag.Bool("v", false, "Optional.  Turns on verbose mode.")
	quiet := flag.Bool("q", false, "Optional.  Turns on quiet mode.")
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

	if !*quiet {
		fmt.Printf("[INFO] Number of files scanned = %d\n", ss.TotalFiles)
		fmt.Printf("[INFO] Number of links matched = %d\n", len(ss.Links))
		fmt.Printf("[INFO] Links skipped = %d\n", len(ss.Skipped))
		fmt.Printf("[INFO] Links failed  = %d\n", len(ss.Failed))
	}

	if *verbose {
		if len(ss.Links) > 0 {
			fmt.Println("\n[DEBUG] -----------------------------------------------------------------------------")
			fmt.Println("[DEBUG] Matched links:")
			fmt.Println("[DEBUG] -----------------------------------------------------------------------------")
			printLinks(ss.Links)
			fmt.Println("[DEBUG] -----------------------------------------------------------------------------")
		} else {
			fmt.Println("[DEBUG] No matched links.")
		}

		if len(ss.Skipped) > 0 {
			fmt.Println("[DEBUG] -----------------------------------------------------------------------------")
			l := strings.Split(ls.SkipPattern, "|")
			s := strings.Join(l, ",")
			fmt.Printf("[DEBUG] Skip pattern list: %s\n", strings.ReplaceAll(s, "\\", ""))
			fmt.Println("[DEBUG] Skipped links:")
			fmt.Println("[DEBUG] -----------------------------------------------------------------------------")
			printLinks(ss.Skipped)
			fmt.Println("[DEBUG] -----------------------------------------------------------------------------")
			fmt.Println()
		} else {
			fmt.Println("[DEBUG] No skipped links.")
		}
	}

	if len(ss.Failed) > 0 {
		if !*quiet {
			for _, link := range ss.Failed {
				var e string
				var ok bool
				if e, ok = statusCodes[link.StatusCode]; !ok {
					e = "Unknown error"
				}
				fmt.Printf("\n(Link)  %s\n(Error) %s\n(Owner) %s\n", link.URL, e, link.Owner)
			}
		}

		if len(ss.Skipped) != len(ss.Failed) {
			os.Exit(1)
		}
	} else {
		if !*quiet {
			fmt.Println("[SUCCESS] No failures.")
		}
	}
}
