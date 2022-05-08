/*
  This software is released under the MIT Licence
  XREF: https://github.com/extradiable/golang-samples
*/

package main

import (
	"bytes"
	"context"
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
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

var mutex sync.Mutex
var cache map[string]int64 = make(map[string]int64)

const (
	DEFAULT_PORT = 80
)

type MessageType int

const (
	INFO_MSG MessageType = iota
	WARNING_MSG
	ERROR_MSG
)

type CtxKey int

const (
	CTX_REQUEST_GUID CtxKey = iota
)

type CustomResponseWriter struct {
	http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
}

func (crw *CustomResponseWriter) Write(p []byte) (int, error) {
	return crw.buf.Write(p)
}

func (crw *CustomResponseWriter) WriteHeader(statusCode int) {
	crw.statusCode = statusCode
}

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
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

func (t MessageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

type Response struct {
	Response interface{} `json:"response,omitempty"`
	Messages []Message   `json:"messages,omitempty"`
}

var fatalResponse []byte

func (r *Response) AddErrorMessage(m string) {
	msg := Message{
		Type:    ERROR_MSG,
		Message: m,
	}
	r.Messages = append(r.Messages, msg)
}

func authenticateUser() {
	fmt.Println("authenticateUser is not implemented")
}

type IlegalArgument struct {
	argument interface{}
	reason   string
}

func (err IlegalArgument) Error() string {
	if err.reason != "" {
		return fmt.Sprintf("ilegal argument: %v. reason: %s", err.argument, err.reason)
	}
	return fmt.Sprintf("ilegal argument: %v", err.argument)
}

func NewIlegalArgument(argument interface{}, reason string) error {
	return IlegalArgument{
		argument: argument,
		reason:   reason,
	}
}

func factorial(n int) (int64, error) {
	var err error
	if n < 0 {
		err = NewIlegalArgument(n, "negative numbers are not allowed")
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

func writeResponse(w http.ResponseWriter, r *http.Request, statusCode int, rsp Response) {
	w.WriteHeader(statusCode)
	b, err := json.MarshalIndent(rsp, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		serverLog(r.Context()).WithError(err).Error("could not marshal response object")
		b = fatalResponse
	}
	w.Write(b)
}

func handleError(statusCode int, response Response, err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			serverLog(r.Context()).WithError(err).Error(response.Messages)
		} else {
			serverLog(r.Context()).Error(response.Messages)
		}
		writeResponse(w, r, statusCode, response)
	})
}

func factorialHdl(w http.ResponseWriter, r *http.Request) {
	var response Response
	v := mux.Vars(r)["number"]
	n, err := strconv.Atoi(v)
	if err != nil {
		switch {
		case errors.Is(err, strconv.ErrRange):
			response.AddErrorMessage(fmt.Sprintf("Bad parameter: number '%s' is out of range", v))
		default:
			response.AddErrorMessage(fmt.Sprintf("Bad parameter: '%s' is not a number", v))
		}
		handleError(http.StatusBadRequest, response, err).ServeHTTP(w, r)
		return
	}
	serverLog(r.Context()).Infof("Computing factorial(%d)", n)
	result, err := factorial(n)
	if err != nil {
		response.AddErrorMessage(err.Error())
		switch {
		case errors.As(err, &IlegalArgument{}):
			handleError(http.StatusBadRequest, response, nil).ServeHTTP(w, r)
		default:
			handleError(http.StatusInternalServerError, response, nil).ServeHTTP(w, r)
		}
		return
	}
	response.Response = result
	writeResponse(w, r, http.StatusOK, response)
	saveCache(v, result)
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
			serverLog(r.Context()).Infof("Returning cached value for %s", k)
			response := Response{
				Response: v,
				Messages: []Message{{
					Type:    INFO_MSG,
					Message: "returning cached value",
				}},
			}
			writeResponse(w, r, http.StatusOK, response)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func authnHdl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverLog(r.Context()).Info("Authenticate")
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
		serverLog(r.Context()).Info("logging request")
		crw := &CustomResponseWriter{
			ResponseWriter: w,
			buf:            &bytes.Buffer{},
			statusCode:     200,
		}
		h.ServeHTTP(crw, r)
		serverLog(r.Context()).Info("logging response")
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(crw.statusCode)
		if _, err := io.Copy(w, crw.buf); err != nil {
			serverLog(r.Context()).WithError(err).Error("Failed to send out response")
		}
	})
}

func metricsHdl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		serverLog(r.Context()).Debugf("Processing took %v", time.Since(start))
	})
}

func tagIdHdl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, setGUID(r))
	})
}

func getGUID(ctx context.Context) string {
	value := ctx.Value(CTX_REQUEST_GUID)
	guid, ok := value.(string)
	if !ok {
		log.Warning("key CTX_REQUEST_GUID was not found on context")
	}
	return guid
}

func setGUID(r *http.Request) *http.Request {
	guid := xid.New()
	ctx := context.WithValue(r.Context(), CTX_REQUEST_GUID, guid.String())
	return r.WithContext(ctx)
}

func serverLog(ctx context.Context) *log.Entry {
	guid := getGUID(ctx)
	entry := log.WithField("guid", guid)
	return entry
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(false)
	var err error
	fatalRsp := Response{
		Messages: []Message{{
			Type:    ERROR_MSG,
			Message: "could not process response",
		}},
	}
	fatalResponse, err = json.MarshalIndent(fatalRsp, "", "")
	if err != nil {
		log.WithError(err).Fatal("Could not prepare fatal response")
	}
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
	chain := alice.New(tagIdHdl, metricsHdl, loggingHdl, authnHdl, cacheHdl).Then(http.HandlerFunc(factorialHdl))
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
