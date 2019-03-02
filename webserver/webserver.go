package webserver

import (
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
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
}

/*
 * Exchange format for HTTP responses.
 */
type HttpResponse struct {
	Header map[string]string
	Body   []byte
}

/*
 * Data structure holding channels for communication between a CGI and the web
 * server.
 */
type WebChannels struct {
	Requests  chan HttpRequest
	Responses chan HttpResponse
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
}

/*
 * Data structure holding the web server's internal state.
 */
type webServerStruct struct {
	cgis   map[string]WebChannels
	config Config
}

/*
 * The public interface of the web server.
 */
type WebServer interface {
	RegisterCgi(path string) WebChannels
	GetCgis() []string
	RemoveCgi(path string)
	Run()
}

/*
 * Set default headers for HTTP(S) responses so that we don't have to set them
 * in every handler. This sets a name for the server, a default MIME type, and
 * disables all forms of caching (local and via proxies).
 */
func (this *webServerStruct) setDefaultHeaders(writer http.ResponseWriter, request *http.Request) {
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
 * A handler for CGI requests.
 */
func (this *webServerStruct) cgiHandler(writer http.ResponseWriter, request *http.Request) {
	request.ParseMultipartForm(REQUEST_SIZE)

	/*
	 * The parsed HTTP request.
	 */
	hrequest := HttpRequest{
		Protocol: request.Proto,
		Method:   request.Method,
		Path:     request.URL.Path,
		Host:     request.Host,
		Params:   map[string]string{},
		Files:    map[string][]multipart.File{},
	}

	/*
	 * Iterate over all form values and parse parameters.
	 */
	for key, values := range request.Form {
		params := strings.Join(values, ",")
		hrequest.Params[key] = params
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
			params := strings.Join(values, ",")
			hrequest.Params[key] = params
		}

		multipartFormFile := multipartForm.File

		/*
		 * Iterate over files in multipart form.
		 */
		for key, handles := range multipartFormFile {
			files := hrequest.Files[key]

			/*
			 * If no slice is present under this key, create one.
			 */
			if files == nil {
				files = []multipart.File{}
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
						files = append(files, fd)
					}

				}

			}

			hrequest.Files[key] = files
		}

	}

	/*
	 * Interact with the CGI via channels to send request, fetch response.
	 */
	cgi := this.cgis[hrequest.Path]
	cgi.Requests <- hrequest
	response := <-cgi.Responses
	this.setDefaultHeaders(writer, request)
	hdr := writer.Header()

	/*
	 * Write response headers.
	 */
	for key, value := range response.Header {
		hdr.Set(key, value)
	}

	writer.Write(response.Body)
}

/*
 * A handler for file requests. This allows, e. g. (X)HTML, CSS, JavaScript
 * content and images to be served.
 */
func (this *webServerStruct) fileHandler(writer http.ResponseWriter, request *http.Request) {
	url := request.URL.Path
	this.setDefaultHeaders(writer, request)
	cfg := this.config

	/*
	 * If navigated to web root, redirect to index file, otherwise serve file.
	 */
	if (url == "") || (url == "/") {
		hdr := writer.Header()
		hdr.Set("Location", cfg.Index)
		writer.WriteHeader(http.StatusFound)
	} else {
		dotPos := strings.LastIndex(url, ".")
		extension := ""

		/*
		 * Check for file extension.
		 */
		if dotPos != -1 {
			dotPosInc := dotPos + 1
			extension = url[dotPosInc:]
		}

		mimetype, present := cfg.MimeTypes[extension]

		/*
		 * Check if a MIME type is registered for this extension.
		 */
		if !present {
			mimetype = cfg.DefaultMime
		}

		path := cfg.WebRoot + url
		fd, err := os.Open(path)
		hdr := writer.Header()

		/*
		 * Check if file exists in web root.
		 */
		if err != nil {
			hdr.Set("Content-type", cfg.ErrorMime)
			fmt.Fprintf(writer, "[ERROR] - '%s' does not exist!\n", url)
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
	this.setDefaultHeaders(writer, request)
	uri := request.RequestURI
	uriChars := []rune(uri)

	/*
	 * Ensure that the URI starts with a slash.
	 */
	if string(uriChars[0]) != "/" {
		uri = "/" + uri
	}

	url := fmt.Sprintf("https://%s:%s%s", host, this.config.TLSPort, uri)
	http.Redirect(writer, request, url, http.StatusFound)
}

/*
 * Registers a CGI with the web server. The 'path' given specifies the URL
 * under which the CGI is available. When the CGI is called, the web server
 * generates a WebRequest and puts it into the request queue.
 */
func (this *webServerStruct) RegisterCgi(path string) WebChannels {
	requests := make(chan HttpRequest)
	responses := make(chan HttpResponse)
	channels := WebChannels{Requests: requests, Responses: responses}

	/*
	 * If no CGI map exists, create one.
	 */
	if this.cgis == nil {
		this.cgis = make(map[string]WebChannels)
	}

	this.cgis[path] = channels
	return channels
}

/*
 * Returns a list of the URLs of all currently registered CGIs.
 */
func (this *webServerStruct) GetCgis() []string {
	cgis := []string{}

	/*
	 * Append all CGI paths to list.
	 */
	for path, _ := range this.cgis {
		cgis = append(cgis, path)
	}

	return cgis
}

/*
 * Remove all CGIs currently registered with the web server.
 */
func (this *webServerStruct) RemoveCgi(path string) {
	delete(this.cgis, path)
}

/*
 * The main function of the web server. This loads the web server configuration
 * from the file system, sets up the HTTP request handlers and runs the HTTP
 * listener.
 */
func (this *webServerStruct) Run() {

	/*
	 * Use only GCM (no CBC) and only SHA-2 (no SHA-1!).
	 */
	ciphersuites := []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
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
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: curves,
		CipherSuites:     ciphersuites,
	}

	cfg := this.config
	tlsAddr := fmt.Sprintf(":%s", cfg.TLSPort)

	/*
	 * The TLS server.
	 */
	tlsServer := http.Server{
		Addr:      tlsAddr,
		TLSConfig: &tlsConfig,
	}

	/*
	 * Register all CGI paths to HTTP handler.
	 */
	for path, _ := range this.cgis {
		http.HandleFunc(path, this.cgiHandler)
	}

	http.HandleFunc("/", this.fileHandler)
	go tlsServer.ListenAndServeTLS(cfg.TLSPublicKey, cfg.TLSPrivateKey)
	httpAddr := fmt.Sprintf(":%s", cfg.Port)
	go http.ListenAndServe(httpAddr, http.HandlerFunc(this.redirect))
}

/*
 * Creates a new web server.
 */
func CreateWebServer(cfg Config) WebServer {
	server := webServerStruct{config: cfg}
	return &server
}
