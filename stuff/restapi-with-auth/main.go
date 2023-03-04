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
var cache map[string]CacheEntry = make(map[string]CacheEntry)

const (
	DEFAULT_PORT    = 80
	DEFAULT_SM_PORT = 808
)

type MessageType int

const (
	INFO_MSG MessageType = iota
	WARNING_MSG
	ERROR_MSG
)

type CtxGUIDKey int

const (
	CTX_REQUEST_GUID CtxGUIDKey = iota
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

type CacheEntry struct {
	Timestamp time.Time
	Value     interface{}
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
var timeoutResponse []byte

func (r *Response) AddMessage(t MessageType, msg string) {
	switch t {
	case ERROR_MSG:
		r.AddErrorMessage(msg)
	case WARNING_MSG:
		r.AddWarningMessage(msg)
	case INFO_MSG:
		r.AddWarningMessage(msg)
	}
}

func (r *Response) AddErrorMessage(m string) {
	msg := Message{
		Type:    ERROR_MSG,
		Message: m,
	}
	r.Messages = append(r.Messages, msg)
}

func (r *Response) AddWarningMessage(m string) {
	msg := Message{
		Type:    WARNING_MSG,
		Message: m,
	}
	r.Messages = append(r.Messages, msg)
}

func (r *Response) AddInfoMessage(m string) {
	msg := Message{
		Type:    INFO_MSG,
		Message: m,
	}
	r.Messages = append(r.Messages, msg)
}

func authenticateUser() {
	fmt.Println("authenticateUser is not implemented")
}

func collatz(n int64) (int, error) {
	if n < 0 {
		return -1, fmt.Errorf("number has to be greater than zero")
	}
	for i := 1; ; i++ {
		if n == 1 {
			return i, nil
		}
		if n%2 == 0 {
			n = n / 2
		} else {
			n = 1 + n*3
			if n < 0 {
				return 0, fmt.Errorf("integer overflow")
			}
		}
	}
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

func formatErrorMessage(err error) string {
	switch {
	case errors.Is(err, strconv.ErrSyntax):
		nerr := err.(*strconv.NumError)
		return fmt.Sprintf("parameter '%s' is not a number", nerr.Num)
	case errors.Is(err, strconv.ErrRange):
		nerr := err.(*strconv.NumError)
		return fmt.Sprintf("got out of range error while processing input '%s'", nerr.Num)
	default:
		return err.Error()
	}
}

func handleError(statusCode int, response Response, err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.AddErrorMessage(formatErrorMessage(err))
		if err != nil {
			serverLog(r.Context()).WithError(err).Error(response.Messages)
		} else {
			serverLog(r.Context()).Error(response.Messages)
		}
		writeResponse(w, r, statusCode, response)
	})
}

func minmaxcollatzHdl(w http.ResponseWriter, r *http.Request) {
	var response Response
	var max int
	var argMax int64
	ctx := r.Context()
	v := mux.Vars(r)["number"]
	n, err := strconv.Atoi(v)
	if err != nil {
		handleError(http.StatusBadRequest, response, err).ServeHTTP(w, r)
		return
	}
	if n <= 0 {
		err = fmt.Errorf("number has to be greater than zero")
		handleError(http.StatusBadRequest, response, err).ServeHTTP(w, r)
		return
	}
	serverLog(r.Context()).Infof("Computing collatz(%d)", n)
	var i int64
	for i = 1; i <= int64(n); i++ {
		select {
		case <-ctx.Done():
			serverLog(r.Context()).WithError(ctx.Err()).Warning("Processig was stopped")
			return
		default:
			val, err := collatz(i)
			if err != nil {
				break
			}
			if val > max {
				max = val
				argMax = i
			}
		}
	}
	if err != nil {
		handleError(http.StatusInternalServerError, response, err).ServeHTTP(w, r)
		return
	}
	response.Response = struct {
		Max    int   `json:"max"`
		Number int64 `json:"number"`
	}{
		Max:    max,
		Number: argMax,
	}
	writeResponse(w, r, http.StatusOK, response)
	saveCache(v, response.Response)
}

func saveCache(k string, v interface{}) {
	log.Debugf("Added key '%s' to cache", k)
	mutex.Lock()
	defer mutex.Unlock()
	cache[k] = CacheEntry{
		Timestamp: time.Now(),
		Value:     v,
	}
}

func queryCache(k string) (interface{}, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	entry, ok := cache[k]
	if ok {
		entry.Timestamp = time.Now()
		cache[k] = entry
	}
	return entry.Value, ok
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
			statusCode:     http.StatusOK,
		}
		h.ServeHTTP(crw, r)
		serverLog(r.Context()).Info("logging response")
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(crw.statusCode)
		if _, err := io.Copy(w, crw.buf); err != nil {
			serverLog(r.Context()).WithError(err).Error("Failed to send out response")
		}
		serverLog(r.Context()).Debug("loggingHdl completed")
	})
}

func metricsHdl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		serverLog(r.Context()).Debugf("Processing took %v", time.Since(start))
	})
}

func timeoutHdl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelCtx := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancelCtx()
		done := make(chan struct{})
		panicChan := make(chan any, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			h.ServeHTTP(w, r.WithContext(ctx))
			close(done)
		}()
		select {
		case p := <-panicChan:
			serverLog(ctx).Warn("Panic caught by timeoutHdl")
			panic(p)
		case <-done:
			switch err := ctx.Err(); err {
			case context.DeadlineExceeded:
				serverLog(ctx).Warn("Deadline Exceeded")
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write(timeoutResponse)
			default:
				serverLog(ctx).Info("timeoutHdl completed in time")
			}
		}
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
	fatalResponse, err = json.Marshal(fatalRsp)
	if err != nil {
		log.WithError(err).Fatal("Could not prepare fatal response")
	}
	timeoutRsp := Response{
		Messages: []Message{{
			Type:    ERROR_MSG,
			Message: "deadline exceeded",
		}},
	}
	timeoutResponse, err = json.Marshal(timeoutRsp)
	if err != nil {
		log.WithError(err).Fatal("Could not prepare timeout response")
	}
}

func monitorCache() {
	for {
		select {
		case <-time.After(time.Duration(30) * time.Second):
			log.Info("Clean cache")
			mutex.Lock()
			for key, entry := range cache {
				if time.Since(entry.Timestamp) > 60*time.Second {
					log.Debugf("Key '%s' was deleted from cache", key)
					delete(cache, key)
				}
			}
			mutex.Unlock()
		}
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
	chain := alice.New(tagIdHdl, metricsHdl, loggingHdl, timeoutHdl, authnHdl, cacheHdl)
	router.Handle("/minmaxcollatz/{number}", chain.Then(http.HandlerFunc(minmaxcollatzHdl)))

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("0.0.0.0:%d", serverPort),
		WriteTimeout: 0,
		ReadTimeout:  5 * time.Second,
		TLSConfig:    tlsConfig,
	}

	go monitorCache()

	log.Fatal(srv.ListenAndServe())
}
