package httpserver

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"
)

func responseHeader(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Date", time.Now().Format(time.RFC1123))
	rw, ok := w.(*responseWriter)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Request-Id", rw.requestID)
	w.Header().Set("X-Request-Id", rw.xRequestID)
	w.WriteHeader(statusCode)
}

// Response response by writing the body to http.ResponseWriter.
// Call at the end line of your handler.
func Response(w http.ResponseWriter, statusCode int, body []byte) {
	responseHeader(w, statusCode)
	w.Write(body)
}

// ResponseJSON response by writing body with json encoder into http.ResponseWriter.
// Body must be either struct or map[string]interface{}. Otherwise would result in incorrect parsing at client side.
// If you have []byte as response body, then use Response function instead.
// Call at the end line of your handler.
func ResponseJSON(w http.ResponseWriter, statusCode int, body interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	responseHeader(w, statusCode)
	return json.NewEncoder(w).Encode(body)
}

// ResponseString response in form of string whatever passed into body param.
// Call at the end line of your handler.
func ResponseString(w http.ResponseWriter, statusCode int, body interface{}) {
	responseHeader(w, statusCode)
	fmt.Fprintf(w, "%v", body)
}

// ResponseXML response by writing body with xml encoder into http.ResponseWriter.
// Body must be either struct or map[string]interface{}. Otherwise would result in incorrect parsing at client side.
// If you have []byte as response body, then use Response function instead.
// Call at the end line of your handler.
func ResponseXML(w http.ResponseWriter, statusCode int, body interface{}) error {
	w.Header().Set("Content-Type", "application/xml")
	responseHeader(w, statusCode)
	return xml.NewEncoder(w).Encode(body)
}

// ResponseHTML render and return html with given data.
// @tmplName: template name if a template is wrapped inside {{ define "tmplName" }}, otherwise empty string.
// @tmpl: template content in form of string loaded from template file.
// @data: data to be embedded into html template, preferably in form of map[string]interface{}.
// @funcMap: golang template FuncMap.
func ResponseHTML(w http.ResponseWriter, tmplName string, tmpl string, data interface{},
	funcMap ...template.FuncMap) error {

	html, err := RenderHTML(tmplName, tmpl, data, funcMap...)
	if err != nil {
		return err
	}
	ResponseString(w, http.StatusOK, html)
	return nil
}

// RenderHTML render template with given data into string.
// @tmplName: template name if a template is wrapped inside {{ define "tmplName" }}, otherwise empty string.
// @tmpl: template content in form of string loaded from template file.
// @data: data to be embedded into html template, preferably in form of map[string]interface{}.
// @funcMap: golang template FuncMap.
func RenderHTML(tmplName string, tmpl string, data interface{}, funcMap ...template.FuncMap) (template.HTML, error) {
	var (
		t    *template.Template
		buff bytes.Buffer
		err  error
	)

	t = template.New(tmplName)
	for _, v := range funcMap {
		t = t.Funcs(v)
	}

	t, err = t.Parse(tmpl)
	if err != nil {
		return "", err
	}

	if err = t.Execute(&buff, data); err != nil {
		return "", err
	}

	return template.HTML(buff.String()), nil
}

func ResponseMultiHTML(w http.ResponseWriter, mainTmplName string, tmplNameToTmpl map[string]string, data interface{}, funcMap ...template.FuncMap) error {
	html, err := RenderMultiHTML(mainTmplName, tmplNameToTmpl, data, funcMap...)
	if err != nil {
		return err
	}
	ResponseString(w, http.StatusOK, html)
	return nil
}

func RenderMultiHTML(mainTmplName string, tmplNameToTmpl map[string]string, data interface{}, funcMap ...template.FuncMap) (template.HTML, error) {
	var (
		t    *template.Template
		buff bytes.Buffer
		err  error
	)

	t = template.New(mainTmplName)
	for _, v := range funcMap {
		t = t.Funcs(v)
	}
	t, err = t.Parse(tmplNameToTmpl[mainTmplName])
	if err != nil {
		return "", err
	}

	for k, v := range tmplNameToTmpl {
		if k != mainTmplName {
			t, err = t.New(k).Parse(v)
			if err != nil {
				return "", err
			}
		}
	}

	if err = t.ExecuteTemplate(&buff, mainTmplName, data); err != nil {
		return "", err
	}

	return template.HTML(buff.String()), nil
}

func LoadTemplate(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
