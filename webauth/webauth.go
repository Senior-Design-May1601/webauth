package main

import (
    "flag"
    "html/template"
    "net/http"
    "strconv"
    "time"
)

type Info struct {
    Failure bool
}

const (
    DEFAULT_CERT = "../tls/dummy_cert.pem"
    DEFAULT_KEY = "../tls/dummy_key.pem"
    LOGIN_TEMPLATE = "login.html"
    LOGIN_URL = "/login/"
)

var templates = template.Must(template.ParseFiles(LOGIN_TEMPLATE))

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

func main() {
    http_str := flag.Int("http", 8080, "HTTP server port")
    https_str := flag.Int("https", 8443, "HTTPS server port")
    cert := flag.String("cert", DEFAULT_CERT, "path to TLS certificate")
    key := flag.String("key", DEFAULT_KEY, "path to TLS private key")
    flag.Parse()

    http_port := ":" + strconv.Itoa(*http_str)
    https_port := ":" + strconv.Itoa(*https_str)

    http.HandleFunc(LOGIN_URL, loginBaseHandler)
    http.HandleFunc("/", redirectHandler)

    go func() {
        err := http.ListenAndServe(http_port, nil)
        if err != nil {
            panic("HTTP server error: " + err.Error())
        }
    }()

    err := http.ListenAndServeTLS(https_port, *cert, *key, nil)
    if err != nil {
        panic("HTTPS server error: " + err.Error())
    }
}
