/*
  This software is released under the MIT Licence
  XREF: https://github.com/extradiable/golang-samples
*/

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
)

var mutex sync.Mutex
var cache map[string]int64 = make(map[string]int64)

const (
	DEFAULT_PORT     = 80
	CTX_REQUEST_GUID = "ctx_request_id"
)

type CustomResponseWriter struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

type MessageType int

const (
	INFO_MSG MessageType = iota
	WARNING_MSG
	ERROR_MSG
)

func (t MessageType) String() string {
	switch t {
	case INFO_MSG:
		return "INFO"
	case WARNING_MSG:
		return "WARNING"
	case ERROR_MSG:
		return "ERROR"
	default:
		return "UNDEFINED"
	}
}

type Message struct {
	Type    MessageType
	Message string
}

type Response struct {
	Response interface{}
	Messages []Message
}

func (crw *CustomResponseWriter) Write(p []byte) (int, error) {
	return crw.buf.Write(p)
}

func authenticateUser() {
	fmt.Println("authenticateUser is not implemented")
}

type IlegalArgument struct {
	argument interface{}
}

func (err IlegalArgument) Error() string {
	return fmt.Sprintf("Ilegal Argument: %v", err.argument)
}

func factorial(n int) (int64, error) {
	var err error
	if n < 0 {
		err = IlegalArgument{
			argument: n,
		}
		return -1, err
	}
	var r int64 = 1
	for i := 1; i <= n; i++ {
		r = r * int64(i)
		if r <= 0 {
			return 0, fmt.Errorf("integer overflow")
		}
	}
	return r, nil
}

func handleError(status int, message string, err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			log.WithError(err).Error(message)
		} else {
			log.Error(message)
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(message))
	})
}

func factorialHdl(w http.ResponseWriter, r *http.Request) {
	var response Response
	v := mux.Vars(r)["number"]
	n, err := strconv.Atoi(v)
	if err != nil {
		message := fmt.Sprintf("Bad parameter: '%s' is not a number", v)
		handleError(http.StatusBadRequest, message, err).ServeHTTP(w, r)
		return
	}
	log.Infof("Computing factorial(%d)", n)
	result, err := factorial(n)
	var ilegalArg *IlegalArgument
	if err != nil {
		switch {
		case errors.As(err, &ilegalArg):
			handleError(http.StatusBadRequest, err.Error(), nil).ServeHTTP(w, r)
		default:
			handleError(http.StatusInternalServerError, err.Error(), nil).ServeHTTP(w, r)
		}
		return
	}
	response.Response = result
	b, err := json.MarshalIndent(response, "", " ")
	if err != nil {
		handleError(http.StatusInternalServerError, "Could not selialize response", err).ServeHTTP(w, r)
	}
	saveCache(v, result)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func saveCache(k string, v int64) {
	mutex.Lock()
	defer mutex.Unlock()
	cache[k] = v
}

func queryCache(k string) (int64, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	v, ok := cache[k]
	return v, ok
}

func cacheHdl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		k := mux.Vars(r)["number"]
		v, ok := queryCache(k)
		if ok {
			log.Infof("Returning cached value for %s", k)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Factorial is %d", v)))
			return
		}
		h.ServeHTTP(w, r)
	})
}

func authnHdl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("Authenticate")
		h.ServeHTTP(w, r)
		/*
			realm := "myRealm"
			domain := "/products"
			headerValue := fmt.Sprintf("realm=\"%s\" domain=\"%s\"", realm, domain)
			r.Header.Add("WWW-Authenticate", headerValue)*/
	})
}

func loggingHdl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("logging request")
		crw := &CustomResponseWriter{
			ResponseWriter: w,
			buf:            &bytes.Buffer{},
		}
		h.ServeHTTP(crw, r)
		log.Info("logging response")
		if _, err := io.Copy(w, crw.buf); err != nil {
			log.Printf("Failed to send out response: %v", err)
		}
	})
}

func metricsHdl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		time.Since(start)
	})
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetReportCaller(false)
}

func main() {

	log.Info("Starting Server")

	serverPort := DEFAULT_PORT
	if len(os.Getenv("SERVER_PORT")) != 0 {
		tmp, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
		if err == nil {
			serverPort = tmp
		}
	}
	log.Infof("Server Port: %d", serverPort)

	tlsConfig := &tls.Config{}
	tlsConfig.ClientAuth = tls.NoClientCert
	tlsConfig.NextProtos = []string{"http/1.1"}

	router := mux.NewRouter()
	chain := alice.New(metricsHdl, loggingHdl, authnHdl, cacheHdl).Then(http.HandlerFunc(factorialHdl))
	router.Handle("/factorial/{number}", chain)

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("0.0.0.0:%d", serverPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		TLSConfig:    tlsConfig,
	}

	log.Fatal(srv.ListenAndServe())
}

/*
func getGUID(r *http.Request) string {
	ctx := r.Context()
	value := ctx.Value(CTX_REQUEST_GUID)
	guid, ok := value.(string)
	if !ok {
		panic("Could not get the GUID from context")
	}
	return guid
}

func setGUID(r *http.Request) *http.Request {
	guid := xid.New()
	ctx := context.WithValue(r.Context(), CTX_REQUEST_GUID, guid.String())
	return r.WithContext(ctx)
}

func serverLog(ctx context.Context) *log.Entry {
	guid := ctx.Value(CTX_REQUEST_GUID).(string)
	entry := log.WithField("guid", guid)
	return entry
}
*/
