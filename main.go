package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Senior-Design-May1601/Splunk/alert"
	"github.com/Senior-Design-May1601/projectmain/logger"
)

type Config struct {
	Host          string
	Key           string
	Cert          string
	HttpPort      int
	HttpsPort     int
	LoginTemplate string
}

type Info struct {
	Failure bool
}

const (
	LOGIN_URL = "/login/"
)

func makeAlert(r *http.Request) string {
	meta := make(map[string]string)

	meta["service"] = "web"
	meta["remote"] = r.RemoteAddr
	meta["local"] = r.Host
	meta["url"] = r.URL.String()
	meta["user-agent"] = r.Header.Get("User-Agent")
	meta["cookie"] = r.Header.Get("Cookie")
	meta["referer"] = r.Header.Get("Referer")
	meta["method"] = r.Method
	err := r.ParseForm()
	if err == nil {
		meta["username"] = r.Form.Get("username")
		meta["password"] = r.Form.Get("password")
	}

	return alert.NewSplunkAlertMessage(meta)
}

func loginBaseHandler(w http.ResponseWriter, r *http.Request) {
	mylogger.Println(makeAlert(r))
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
	templates.ExecuteTemplate(w, loginTemplate, &Info{Failure: false})
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: should we throttle this a bit to simulate DB call?
	setHeader(w)
	templates.ExecuteTemplate(w, loginTemplate, &Info{Failure: true})
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	setHeader(w)
	mylogger.Println(makeAlert(r))
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

var templates *template.Template
var loginTemplate string
var config Config
var mylogger *log.Logger

func main() {
	mylogger = logger.NewLogger("", 0)

	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	if _, err := toml.DecodeFile(*configPath, &config); err != nil {
		mylogger.Fatal(err)
	}

	s := strings.Split(config.LoginTemplate, "/")
	loginTemplate = s[len(s)-1]

	http.HandleFunc("/login/", loginBaseHandler)
	http.HandleFunc("/", redirectHandler)

	templates = template.Must(template.ParseFiles(config.LoginTemplate))

	go func() {
		err := http.ListenAndServe(config.Host+":"+strconv.Itoa(config.HttpPort), nil)
		if err != nil {
			mylogger.Fatal("Http server error: " + err.Error())
		}
	}()

	err := http.ListenAndServeTLS(config.Host+":"+strconv.Itoa(config.HttpsPort),
		config.Cert,
		config.Key,
		nil)
	if err != nil {
		mylogger.Fatal("Https server error: " + err.Error())
	}
}
