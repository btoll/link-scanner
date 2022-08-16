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
	405: "405 Method Not Allowed",
	429: "429 Too Many Requests",
	500: "500 Internal Server Error",
	501: "501 Not Implemented",
	502: "502 Bad Gateway",
	503: "503 Service Unavailable",
}

func getHead(c chan Link, l Link, headers http.Header) {
	/*
		_, err := url.ParseRequestURI(l.URL)
		if err != nil {
			l.StatusCode = -1
			l.Error = err
			c <- l
			return
		}
	*/

	client := http.Client{}
	req, err := http.NewRequest("HEAD", l.URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header = headers
	resp, err := client.Do(req)
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

func getHeaders(headers string) http.Header {
	// Set some sensible defaults.
	h := http.Header{
		"Content-Type": {"text/html"},
		"User-Agent":   {"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:103.0) Gecko/20100101 Firefox/103.0"},
	}
	if len(headers) == 0 {
		return h
	}
	for _, header := range strings.Split(headers, ",") {
		// Specify the expected split number because a header key value can contain colons,
		// i.e., a User-Agent value (see default value).
		v := strings.SplitN(header, ":", 2)
		if len(v) != 2 {
			fmt.Println("Bad header values.")
			os.Exit(1)
		}
		// Header.Set will override previous keys.
		h.Set(v[0], v[1])
	}
	return h
}

func main() {
	dir := flag.String("dir", "", "Optional.  Searches every file in the directory for a match.  Non-recursive.")
	filename := flag.String("filename", "", "Optional.  Takes precedence over directory searches.")
	filetype := flag.String("filetype", ".md", "Only searches files of this type.  Include the period, i.e., `.html`")
	header := flag.String("header", "", "Optional.  Takes comma-delimited pairs of key:value")
	verbose := flag.Bool("v", false, "Optional.  Turns on verbose mode.")
	quiet := flag.Bool("q", false, "Optional.  Turns on quiet mode.")
	flag.Parse()

	h := getHeaders(*header)
	ls := LinkScanner{
		Dir:         *dir,
		FileName:    *filename,
		FileType:    *filetype,
		SkipPattern: `\.onion|example\.com`,
		Header:      h,
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
		fmt.Printf("[INFO] Links failed  = %d\n\n", len(ss.Failed))
		fmt.Println("[INFO] Request headers")
		for k, v := range h {
			fmt.Printf("\t%s: %s\n", k, v)
		}
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
					e = fmt.Sprintf("%d Unknown error", link.StatusCode)
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
