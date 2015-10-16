package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"text/template"
)

var mainPage = template.Must(template.New("page").Parse(`
<html>
	<head>
		<style type="text/css">
			h1, h2, h3 {
				line-height: 1.2;
			}
			body {
				margin: 40px auto;
				max-width: 650px;
				line-height: 1.6;
				font-size: 18px;
				color: #444;
				padding: 0 10px;
			}
		</style>
	</head>
	<body>
	  <a href="/"><h1>Home</h1></a>
		<form action="render">
			Url to render: <input type="text" name="url" value="">
			<input type="submit" value="Submit">
		</form> 
		<div>{{.Rendered}}</div>
	</body>
</html>`))

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mainPage.Execute(w, nil)
	})
	http.HandleFunc("/render", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			mainPage.Execute(w, struct{ Rendered string }{"Bad request"})
			return
		}

		if _, err := url.Parse(r.Form.Get("url")); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			mainPage.Execute(w, struct{ Rendered string }{"Bad request"})
			return
		}

		res, err := http.Get(r.Form.Get("url"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			mainPage.Execute(w, struct{ Rendered string }{"Couldn't get remote url"})
			return
		}
		defer res.Body.Close()

		if res.StatusCode > 300 {
			w.WriteHeader(http.StatusInternalServerError)
			mainPage.Execute(w, struct{ Rendered string }{fmt.Sprintf("Error getting page (%d)", res.StatusCode)})
			return
		}

		cmd := exec.Command("groff", "-Thtml", "-mandoc", "-")
		cmd.Stdin = res.Body
		var buf bytes.Buffer
		cmd.Stdout = &buf
		if err := cmd.Run(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			mainPage.Execute(w, struct{ Rendered string }{"Error generating the man page"})
			return
		}

		mainPage.Execute(w, struct{ Rendered string }{buf.String()})
	})
	http.ListenAndServe(":8080", nil)
}
