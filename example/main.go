package main

import (
	"fmt"
	"net/http"
	"os"

	_httpserver "github.com/mfathirirhas/httpserver"
)

// /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.key
// /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.csr
// /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.crt
// openssl req  -new  -newkey rsa:2048  -nodes  -keyout /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.key  -out /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.csr
// openssl  x509  -req  -days 365  -in /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.csr  -signkey /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.key  -out /Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.crt

func main() {
	srv := _httpserver.New(&_httpserver.Opts{
		Port:         8080,
		EnableLogger: true,
		IdleTimeout:  10,
		PanicHandler: func(w http.ResponseWriter, r *http.Request, rcv ...interface{}) {
			fmt.Println("PANIX.................................!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		},
	})
	// srv.TLSConfig(
	// 	"/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.crt",
	// 	"/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/localhost.key",
	// )
	msrv := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("This is middleware for all handlers in this server")
			next(w, r)
		}
	}
	srv.Use(msrv)
	m1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("This middleware m1 : ", os.Getpid())
			next(w, r)
		}
	}
	// serve files inside example/
	srv.FILES("/html/*filepath", "/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/")
	jsFiles := srv.Group("/js")
	jsFiles.FILES("/*filepath", "/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/")
	srv.GET("/handler1", Handler1, m1)
	srv.POST("/handler2", Handler2)
	mg1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("This is group 1 middleware: ", os.Getpid())
			next(w, r)
		}
	}
	g1 := srv.Group("/v1", mg1)
	g1.GET("/handler1", Handler1)
	gh2 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("This is group 1 handler 2 middleware: ", os.Getpid())
			next(w, r)
		}
	}
	g1.POST("/handler2", Handler2, gh2)
	srv.GET("/single-template", HandlerSingleTemplate)
	srv.GET("/multi-template", HandlerMultipleTemplate)
	srv.GET("/multi-template-c", HandlerMultipleTemplateCombined)
	srv.Run()
}

func Handler1(w http.ResponseWriter, r *http.Request) {
	allHeaders := r.Header
	allParams := r.URL.RawQuery
	_httpserver.ResponseString(w, http.StatusOK, fmt.Sprintf("Handler1: %s | %s", allHeaders, allParams))
}

func Handler2(w http.ResponseWriter, r *http.Request) {
	allHeaders := r.Header
	allParams := r.URL.RawQuery
	r.ParseForm()
	b1 := r.FormValue("key1")
	b2 := r.FormValue("key2")
	resp := make(map[string]interface{})
	resp["headers"] = allHeaders
	resp["params"] = allParams
	resp["body1"] = b1
	resp["body2"] = b2
	_httpserver.ResponseJSON(w, http.StatusOK, resp)
}

func HandlerSingleTemplate(w http.ResponseWriter, r *http.Request) {
	htmlFile := "/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/test.html"
	tmpl, _ := _httpserver.LoadTemplate(htmlFile)
	m := map[string]interface{}{
		"this": "printThis",
	}
	fmt.Println(_httpserver.ResponseHTML(w, "test", tmpl, m))
}

func HandlerMultipleTemplate(w http.ResponseWriter, r *http.Request) {
	headerFile := "/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/test_template_header.html"
	headerContent, _ := _httpserver.LoadTemplate(headerFile)
	headerData := map[string]interface{}{
		"pageTitle": "This is Title",
	}
	header, _ := _httpserver.RenderHTML("header", headerContent, headerData)
	m := map[string]interface{}{
		"header": header,
	}

	mainLayout := "/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/test_template.html"
	mainContent, _ := _httpserver.LoadTemplate(mainLayout)
	_httpserver.ResponseHTML(w, "", mainContent, m)
}

func HandlerMultipleTemplateCombined(w http.ResponseWriter, r *http.Request) {
	headerFile := "/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/test_multi_header.html"
	headerContent, _ := _httpserver.LoadTemplate(headerFile)
	footerFile := "/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/test_multi_footer.html"
	footerContent, _ := _httpserver.LoadTemplate(footerFile)
	mainFile := "/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/httpserver/example/test_multi_main.html"
	mainContent, _ := _httpserver.LoadTemplate(mainFile)

	data := map[string]interface{}{
		"pageTitle": "This is Title",
	}
	m := map[string]string{
		"header": headerContent,
		"footer": footerContent,
		"main":   mainContent,
	}

	_httpserver.ResponseMultiHTML(w, "main", m, data)
}
