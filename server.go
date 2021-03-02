package main

import (
	"context"
	"flag"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var (
	logger          = log.New(os.Stdout, "", log.LstdFlags)
	listenAddr      string
	respLengthMax   int
	respUnitDefault int
	respTimeMax     int
)

func randLetterBytes(n int) []byte {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return b
}

func getIntQueryParameterOr(r *http.Request, key string, defaultValue int) int {
	sv := r.URL.Query().Get(key)
	value, err := strconv.Atoi(sv)
	if sv == "" || err != nil {
		return defaultValue
	} else {
		return value
	}
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("| %s (UserAgent: %s) --> %s %s %s %s\n", r.RemoteAddr, r.UserAgent(), r.Host, r.Proto, r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Generate resp parameters
	respLength := getIntQueryParameterOr(r, "length", rand.Intn(respLengthMax)+1) // response size [byte]
	respUnit := getIntQueryParameterOr(r, "unit", respUnitDefault)                // response size unit [byte]
	respTime := getIntQueryParameterOr(r, "time", rand.Intn(respTimeMax)+1)       // response time [sec]
	logger.Printf("| Response: length = %d [byte], unit = %d [byte], time = %d [sec]\n", respLength, respUnit, respTime)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(respLength))
	w.WriteHeader(http.StatusOK)

	n := int(math.Ceil(float64(respLength) / float64(respUnit)))
	t := respTime * 1000 / n
	for i := 0; i < n; i++ {
		time.Sleep(time.Duration(t) * time.Millisecond)
		b := int(math.Min(float64(respLength-(respUnit*i)), float64(respUnit)))
		w.Write(randLetterBytes(b))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// args
	flag.IntVar(&respTimeMax, "time-max", 600, "Maximum response time in [sec] for randomly determining the processing time.")
	flag.IntVar(&respLengthMax, "length-max", 10*1024*1024, "Maximum response body size in for randomly generating response body.")
	flag.IntVar(&respUnitDefault, "unit-default", 1024, "Default number of bytes to be written in a single loop.")
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "Server listen address.")
	flag.Parse()

	// router
	mux := http.NewServeMux()
	mux.HandleFunc("/", loggingMiddleware(indexHandler))

	// http.Server settings
	server := &http.Server{
		Addr:         listenAddr,
		WriteTimeout: time.Second * time.Duration(respTimeMax),
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Second * 30,
		Handler:      mux,
	}

	// start up http.Server with graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Fatalf("%s", err)
		}
	}()

	logger.Println("Server is ready to handle requests at", listenAddr)
	<-sigs

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Println("Failed to gracefully shutdown HTTPServer:", err)
	}
	os.Exit(0)
}
