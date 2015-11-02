package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"net/http"
	"os"
	"time"
)

type Configuration struct {
	Certificate string
	Key         string
	Http_port   string
	Https_port  string
}

type Info struct {
	Failure bool
}

const (
	LOGIN_TEMPLATE = "login.html"
	TEMPLATE_PATH  = "templates/"
	LOGIN_URL      = "/login/"
)

var templates = template.Must(template.ParseFiles(TEMPLATE_PATH +
	LOGIN_TEMPLATE))

func loginBaseHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != LOGIN_URL {
		redirectHandler(w, r)
		return
	}

	switch r.Method {
	case "GET":
		loginGetHandler(w, r)
		break
	case "POST":
		loginPostHandler(w, r)
		break
	default:
		redirectHandler(w, r)
	}
}

func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	setHeader(w)
	templates.ExecuteTemplate(w, LOGIN_TEMPLATE, &Info{Failure: false})
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: should we throttle this a bit to simulate DB call?
	setHeader(w)
	templates.ExecuteTemplate(w, LOGIN_TEMPLATE, &Info{Failure: true})
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	setHeader(w)
	http.Redirect(w, r, LOGIN_URL, 302)
}

func setHeader(w http.ResponseWriter) {
	// reorder the headers so 'Date' comes first to mimic Apache,
	// making sure 'Date' looks like a real date header
	d := time.Now().Format(time.RFC1123)
	l := len(d)
	d = d[:l-4] + " GMT\r\n"
	w.Header().Set("Date", d)
	w.Header().Set("Server", "Apache")
	w.Header().Set("X-Powered-By", "PHP/5.4.41")
}

func read_config(path string) Configuration {
	file, _ := os.Open(path)
	decoder := json.NewDecoder(file)
	config := Configuration{}
	err := decoder.Decode(&config)
	if err != nil {
		panic("Configuration error: " + err.Error())
	}

	return config
}

func main() {
	config_path := flag.String("config", "", "path to config file")
	flag.Parse()
	config := read_config(*config_path)

	http.HandleFunc(LOGIN_URL, loginBaseHandler)
	http.HandleFunc("/", redirectHandler)

	go func() {
		err := http.ListenAndServe(":"+config.Http_port, nil)
		if err != nil {
			panic("HTTP server error: " + err.Error())
		}
	}()

	err := http.ListenAndServeTLS(":"+config.Https_port,
		config.Certificate,
		config.Key, nil)
	if err != nil {
		panic("HTTPS server error: " + err.Error())
	}
}
