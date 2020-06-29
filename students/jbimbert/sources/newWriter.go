package sources

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Here we will replace the original ResponseWriter with our
// which will overwrite some functions we want to achieve our goal
type SrcResponseWriter struct {
	// Header() Header
	// Write([]byte) (int, error)
	// WriteHeader(statusCode int)
	http.ResponseWriter
	writes     [][]byte
	statusCode int
}

func (m *SrcResponseWriter) Write(b []byte) (int, error) {
	m.writes = append(m.writes, b)
	return len(b), nil
}

func (m *SrcResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

// transfer data from myResponseWriter to the original ResponseWriter
func (m *SrcResponseWriter) Flush() {
	if m.statusCode != 0 {
		m.ResponseWriter.WriteHeader(m.statusCode)
	}
	for _, b := range m.writes {
		m.ResponseWriter.Write(b)
	}
}

type UrlNb struct {
	U *url.URL // the url to a file
	N []int    // the line numbers to highlight
}

// Parse the stack trace and return the slice of URLs it contains
func ParseStack(stack []byte) []UrlNb {
	re := regexp.MustCompile(`(/[0-9a-zA-Z/_.]*.go):([0-9]*)`)
	all := re.FindAllSubmatch(stack, -1) // all is of type [][][]byte
	res := make([]UrlNb, 0)
	for _, a := range all { // a of type [][]byte where a[0]=candidate a[1]=u and a[2]=line number
		i, err := strconv.Atoi(string(a[2]))
		if err != nil {
			continue
		}
		u, err := url.Parse(string(a[1]))
		if err != nil {
			continue
		}
		if id := findUrlId(&res, u); id == -1 {
			res = append(res, UrlNb{U: u, N: []int{i}})
		} else {
			res[id].N = append(res[id].N, i)
		}
	}
	return res
}

func findUrlId(urls *[]UrlNb, url *url.URL) int {
	for i, u := range *urls {
		if *u.U == *url {
			return i
		}
	}
	return -1
}

// Replace the stack URLs with the real anchors
func StackToUrls(stack string, urls []UrlNb) string {
	var s string = stack
	for _, u := range urls {
		old := u.U.String()
		// anchor := fmt.Sprintf("<a href=\"/debug?path=%s\">%s</a>", old, old) // Old way to do
		v := url.Values{}
		v.Set("path", old)
		v.Add("lines", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(u.N)), ","), "[]"))
		anchor := fmt.Sprintf("<a href=\"/debug?%s\">%s</a>", v.Encode(), old)
		s = strings.ReplaceAll(s, old, anchor)
	}
	return s
}

//load a file into a slice of strings
func Load(filename string) ([]string, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
