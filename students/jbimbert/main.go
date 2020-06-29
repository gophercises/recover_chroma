package main

import (
	"Gophercizes/recover_chroma/students/jbimbert/sources"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug", handleFile) // Add a function to be called when debugging
	mux.HandleFunc("/panic/", panicDemo)
	mux.HandleFunc("/panic-after/", panicAfterDemo)
	mux.HandleFunc("/", hello)
	log.Fatal(http.ListenAndServe(":3000", devMw(mux)))
}

// Used as http://localhost:3000/debug?path=/home/jimbert/Projects/DiversGo/src/Gophercizes/recover_chroma/students/jbimbert/main.go
// func handleFile(w http.ResponseWriter, r *http.Request) {
// 	p := r.FormValue("path")
// 	file, err := sources.Load(p)
// 	err = quick.Highlight(w, strings.Join(file, "\n"), "go", "html", "dracula")
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// }

// Same a above but with line highlighting
func handleFile(w http.ResponseWriter, r *http.Request) {
	p := r.FormValue("path")
	file, err := sources.Load(p)
	if err != nil {
		log.Println(err)
		return
	}
	lines := r.FormValue("lines")
	numbers := strings.Split(lines, ",")
	linenumbers := make([][2]int, 0)
	for _, sn := range numbers {
		n, _ := strconv.Atoi(sn)
		linenumbers = append(linenumbers, [2]int{n, n})
	}
	lexer := lexers.Get("go")
	it, err := lexer.Tokenise(nil, strings.Join(file, "\n"))
	if err != nil {
		log.Println(err)
		return
	}
	style := styles.Get("dracula")
	if style == nil {
		style = styles.Fallback
	}
	formatter := html.New(html.TabWidth(2), html.WithLineNumbers(true), html.HighlightLines(linenumbers))
	w.Header().Set("Content-Type", "text/html")
	formatter.Format(w, style, it)
}

func devMw(app http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Replace the original ResponseWriter with our myResponseWriter
				mrw := &sources.SrcResponseWriter{ResponseWriter: w}
				log.Println(err)
				stack := debug.Stack()

				urls := sources.ParseStack(stack)                    // Parse the stack and extract the URLs
				urlStack := sources.StackToUrls(string(stack), urls) // Transform the stack paths into anchors of URLs

				log.Println(string(stack))
				mrw.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(mrw, "<h1>panic: %v</h1><pre>%s</pre>", err, urlStack)
				mrw.Flush()
			}
		}()
		app.ServeHTTP(w, r) // Serve the application
	}
}

//************************** original code should no be modified ******************************

func panicDemo(w http.ResponseWriter, r *http.Request) {
	funcThatPanics()
}

func panicAfterDemo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello!</h1>")
	funcThatPanics()
}

func funcThatPanics() {
	panic("Oh no!")
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Hello!</h1>")
}
