package webserver

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	REQUEST_SIZE = 1 << 20
)

/*
 * Exchange format for HTTP requests.
 */
type HttpRequest struct {
	Protocol string
	Method   string
	Path     string
	Host     string
	Params   map[string]string
	Files    map[string][]multipart.File
	Respond  chan<- HttpResponse
}

/*
 * Exchange format for HTTP responses.
 */
type HttpResponse struct {
	Header map[string]string
	Body   []byte
}

/*
 * Data structure representing protocol timeouts.
 *
 * All numbers are provided in seconds.
 *
 * A value of zero represents no timeout.
 */
type ProtocolTimeouts struct {
	Header uint32
	Read   uint32
	Write  uint32
	Idle   uint32
}

/*
 * Data structure representing timeouts for multiple protocols.
 */
type Timeouts struct {
	HTTP ProtocolTimeouts
	TLS  ProtocolTimeouts
}

/*
 * Data structure for web server configuration.
 */
type Config struct {
	Name          string
	Port          string
	TLSPort       string
	TLSPrivateKey string
	TLSPublicKey  string
	WebRoot       string
	Index         string
	MimeTypes     map[string]string
	DefaultMime   string
	ErrorMime     string
	Timeouts      Timeouts
}

/*
 * Data structure holding the web server's internal state.
 */
type webServerStruct struct {
	cgis   map[string]chan<- HttpRequest
	config Config
}

/*
 * The public interface of the web server.
 */
type WebServer interface {
	RegisterCgi(path string) <-chan HttpRequest
	GetCgis() []string
	RemoveCgi(path string)
	Run()
}

/*
 * Set default headers for HTTP(S) responses so that we don't have to set them
 * in every handler. This sets a name for the server, a default MIME type, and
 * disables all forms of caching (local and via proxies).
 */
func (this *webServerStruct) setDefaultHeaders(writer http.ResponseWriter) {
	cfg := this.config
	srv := cfg.Name
	mime := cfg.DefaultMime
	hdr := writer.Header()
	hdr.Set("Server", srv)
	hdr.Set("Content-type", mime)
	hdr.Set("Cache-control", "max-age=0, no-cache, no-store")
	hdr.Set("Pragma", "no-cache")
}

/*
 * Limits the size of an incoming request.
 */
func (this *webServerStruct) limitRequestSize(writer http.ResponseWriter, request *http.Request) {
	requestBody := request.Body
	limitedBody := http.MaxBytesReader(writer, requestBody, REQUEST_SIZE)
	request.Body = limitedBody
}

/*
 * A handler for CGI requests.
 */
func (this *webServerStruct) cgiHandler(writer http.ResponseWriter, request *http.Request) {
	this.limitRequestSize(writer, request)
	request.ParseMultipartForm(REQUEST_SIZE)
	protocol := request.Proto
	method := request.Method
	url := request.URL
	path := url.Path
	host := request.Host
	params := make(map[string]string)
	files := make(map[string][]multipart.File)

	/*
	 * Iterate over all form values and parse parameters.
	 */
	for key, values := range request.Form {
		ps := strings.Join(values, ",")
		params[key] = ps
	}

	multipartForm := request.MultipartForm

	/*
	 * Check if a multipart form is available.
	 */
	if multipartForm != nil {
		multipartFormValue := multipartForm.Value

		/*
		 * Iterate over values in multipart form.
		 */
		for key, values := range multipartFormValue {
			ps := strings.Join(values, ",")
			params[key] = ps
		}

		multipartFormFile := multipartForm.File

		/*
		 * Iterate over files in multipart form.
		 */
		for key, handles := range multipartFormFile {
			fs := files[key]

			/*
			 * If no slice is present under this key, create one.
			 */
			if fs == nil {
				fs = []multipart.File{}
			}

			/*
			 * Iterate over each file handle for this key.
			 */
			for _, handle := range handles {

				/*
				 * Ensure that the handle is not nil.
				 */
				if handle != nil {
					fd, err := handle.Open()

					/*
					 * If the handle points to a file, store file descriptor.
					 */
					if err == nil {
						fs = append(fs, fd)
					}

				}

			}

			files[key] = fs
		}

	}

	responseChannel := make(chan HttpResponse)

	/*
	 * The parsed HTTP request.
	 */
	hrequest := HttpRequest{
		Protocol: protocol,
		Method:   method,
		Path:     path,
		Host:     host,
		Params:   params,
		Files:    files,
		Respond:  responseChannel,
	}

	/*
	 * Interact with the CGI via channels to send request, fetch response.
	 */
	cgis := this.cgis
	cgi := cgis[path]
	cgi <- hrequest
	response := <-responseChannel
	this.setDefaultHeaders(writer)
	hdr := writer.Header()

	/*
	 * Write response headers.
	 */
	for key, value := range response.Header {
		hdr.Set(key, value)
	}

	body := response.Body
	writer.Write(body)
}

/*
 * A handler for file requests. This allows, e. g. (X)HTML, CSS, JavaScript
 * content and images to be served.
 */
func (this *webServerStruct) fileHandler(writer http.ResponseWriter, request *http.Request) {
	url := request.URL
	path := url.Path
	this.setDefaultHeaders(writer)
	cfg := this.config

	/*
	 * If navigated to web root, redirect to index file, otherwise serve file.
	 */
	if (path == "") || (path == "/") {
		hdr := writer.Header()
		indexFile := cfg.Index
		hdr.Set("Location", indexFile)
		writer.WriteHeader(http.StatusFound)
	} else {
		dotPos := strings.LastIndex(path, ".")
		extension := ""

		/*
		 * Check for file extension.
		 */
		if dotPos != -1 {
			dotPosInc := dotPos + 1
			extension = path[dotPosInc:]
		}

		mimetype, present := cfg.MimeTypes[extension]

		/*
		 * Check if a MIME type is registered for this extension.
		 */
		if !present {
			mimetype = cfg.DefaultMime
		}

		webRoot := cfg.WebRoot
		filePath := webRoot + path
		fd, err := os.Open(filePath)
		hdr := writer.Header()

		/*
		 * Check if file exists in web root.
		 */
		if err != nil {
			errorMime := cfg.ErrorMime
			hdr.Set("Content-type", errorMime)
			fmt.Fprintf(writer, "[ERROR] - '%s' does not exist!\n", path)
		} else {
			hdr.Set("Content-type", mimetype)
			io.Copy(writer, fd)
		}

	}

}

/*
 * Redirect insecure requests to TLS.
 */
func (this *webServerStruct) redirect(writer http.ResponseWriter, request *http.Request) {
	split := strings.SplitN(request.Host, ":", 2)
	host := split[0]
	this.setDefaultHeaders(writer)
	uri := request.RequestURI

	/*
	 * Ensure that the URI starts with a slash.
	 */
	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	cfg := this.config
	tlsPort := cfg.TLSPort
	url := fmt.Sprintf("https://%s:%s%s", host, tlsPort, uri)
	http.Redirect(writer, request, url, http.StatusFound)
}

/*
 * Registers a CGI with the web server. The 'path' given specifies the URL
 * under which the CGI is available. When the CGI is called, the web server
 * generates a WebRequest and puts it into the request queue.
 */
func (this *webServerStruct) RegisterCgi(path string) <-chan HttpRequest {
	requests := make(chan HttpRequest)
	cgis := this.cgis

	/*
	 * If no CGI map exists, create one.
	 */
	if cgis == nil {
		cgis = make(map[string]chan<- HttpRequest)
		this.cgis = cgis
	}

	cgis[path] = requests
	return requests
}

/*
 * Returns a list of the URLs of all currently registered CGIs.
 */
func (this *webServerStruct) GetCgis() []string {
	cgis := this.cgis
	cgisNew := []string{}

	/*
	 * Append all CGI paths to list.
	 */
	for path, _ := range cgis {
		cgisNew = append(cgisNew, path)
	}

	return cgisNew
}

/*
 * Remove all CGIs currently registered with the web server.
 */
func (this *webServerStruct) RemoveCgi(path string) {
	cgis := this.cgis
	delete(cgis, path)
}

/*
 * The main function of the web server. This loads the web server configuration
 * from the file system, sets up the HTTP request handlers and runs the HTTP
 * listener.
 */
func (this *webServerStruct) Run() {
	redirectHandler := this.redirect
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", redirectHandler)
	discard := ioutil.Discard
	logger := log.New(discard, "", log.LstdFlags)
	cfg := this.config
	httpPort := cfg.Port
	httpAddr := fmt.Sprintf(":%s", httpPort)
	timeouts := cfg.Timeouts
	httpTimeouts := timeouts.HTTP
	httpTimeoutHeaderSec := httpTimeouts.Header
	httpTimeoutHeaderDur := time.Duration(httpTimeoutHeaderSec)
	httpTimeoutHeader := httpTimeoutHeaderDur * time.Second
	httpTimeoutReadSec := httpTimeouts.Read
	httpTimeoutReadDur := time.Duration(httpTimeoutReadSec)
	httpTimeoutRead := httpTimeoutReadDur * time.Second
	httpTimeoutWriteSec := httpTimeouts.Write
	httpTimeoutWriteDur := time.Duration(httpTimeoutWriteSec)
	httpTimeoutWrite := httpTimeoutWriteDur * time.Second
	httpTimeoutIdleSec := httpTimeouts.Idle
	httpTimeoutIdleDur := time.Duration(httpTimeoutIdleSec)
	httpTimeoutIdle := httpTimeoutIdleDur * time.Second

	/*
	 * The HTTP server.
	 */
	httpServer := http.Server{
		Addr:              httpAddr,
		ErrorLog:          logger,
		Handler:           httpMux,
		IdleTimeout:       httpTimeoutIdle,
		ReadHeaderTimeout: httpTimeoutHeader,
		ReadTimeout:       httpTimeoutRead,
		WriteTimeout:      httpTimeoutWrite,
	}

	tlsMux := http.NewServeMux()
	cgis := this.cgis
	cgiHandler := this.cgiHandler

	/*
	 * Register all CGI paths to TLS server.
	 */
	for path, _ := range cgis {
		tlsMux.HandleFunc(path, cgiHandler)
	}

	fileHandler := this.fileHandler
	tlsMux.HandleFunc("/", fileHandler)
	tlsPort := cfg.TLSPort
	tlsAddr := fmt.Sprintf(":%s", tlsPort)

	/*
	 * TLS cipher suites to use.
	 */
	ciphersuites := []uint16{
		tls.TLS_CHACHA20_POLY1305_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}

	/*
	 * Curves to use for elliptic curve cryptography.
	 */
	curves := []tls.CurveID{
		tls.X25519,
	}

	/*
	 * Use at least TLS 1.2 and Curve25519 (no NIST-Curves!).
	 */
	tlsConfig := tls.Config{
		CipherSuites:             ciphersuites,
		CurvePreferences:         curves,
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
	}

	tlsTimeouts := timeouts.TLS
	tlsTimeoutHeaderSec := tlsTimeouts.Header
	tlsTimeoutHeaderDur := time.Duration(tlsTimeoutHeaderSec)
	tlsTimeoutHeader := tlsTimeoutHeaderDur * time.Second
	tlsTimeoutReadSec := tlsTimeouts.Read
	tlsTimeoutReadDur := time.Duration(tlsTimeoutReadSec)
	tlsTimeoutRead := tlsTimeoutReadDur * time.Second
	tlsTimeoutWriteSec := tlsTimeouts.Write
	tlsTimeoutWriteDur := time.Duration(tlsTimeoutWriteSec)
	tlsTimeoutWrite := tlsTimeoutWriteDur * time.Second
	tlsTimeoutIdleSec := tlsTimeouts.Idle
	tlsTimeoutIdleDur := time.Duration(tlsTimeoutIdleSec)
	tlsTimeoutIdle := tlsTimeoutIdleDur * time.Second

	/*
	 * The TLS server.
	 */
	tlsServer := http.Server{
		Addr:              tlsAddr,
		ErrorLog:          logger,
		Handler:           tlsMux,
		IdleTimeout:       tlsTimeoutIdle,
		ReadHeaderTimeout: tlsTimeoutHeader,
		ReadTimeout:       tlsTimeoutRead,
		TLSConfig:         &tlsConfig,
		WriteTimeout:      tlsTimeoutWrite,
	}

	publicKey := cfg.TLSPublicKey
	privateKey := cfg.TLSPrivateKey
	go tlsServer.ListenAndServeTLS(publicKey, privateKey)
	go httpServer.ListenAndServe()
}

/*
 * Creates a new web server.
 */
func CreateWebServer(cfg Config) WebServer {
	server := webServerStruct{config: cfg}
	return &server
}
