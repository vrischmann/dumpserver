package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/vrischmann/flagutil"
)

type dumper struct {
	logFn  func(format string, args ...interface{})
	output io.Writer
}

func serializeDump(data []byte) []byte {
	var buf bytes.Buffer

	buf.WriteString(time.Now().String())
	buf.WriteString("\n---\n")
	buf.Write(data)
	buf.WriteString("\n---\n\n")

	return buf.Bytes()
}

func (d *dumper) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		w.WriteHeader(200)
	}()

	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Printf("unable to dump request. err=%v", err)
		return
	}

	_, err = d.output.Write(serializeDump(dump))
	if err != nil {
		log.Printf("unable to write request dump. err=%v", err)
	}
}

type humanBytes uint64

func (b *humanBytes) Set(value string) error {
	s, err := humanize.ParseBytes(value)
	if err != nil {
		return err
	}

	*b = humanBytes(s)

	return nil
}

func (b *humanBytes) String() string {
	return humanize.Bytes(uint64(*b))
}

type checkSizeFileWriter struct {
	maxSize int64
	f       *os.File
}

func (w *checkSizeFileWriter) Write(data []byte) (n int, err error) {
	fi, err := w.f.Stat()
	if err != nil {
		return -1, err
	}

	if fi.Size()+int64(len(data)) >= w.maxSize {
		return -1, fmt.Errorf("write of %d bytes would make the file bigger than the max of %s", len(data), humanize.Bytes(uint64(fi.Size())))
	}

	return w.f.Write(data)
}

func main() {
	var (
		flListenAddr    = flag.String("l", flagutil.EnvOrDefault("LISTEN_ADDR", "localhost:4567"), "The listen address")
		flMaxOutputSize humanBytes
		flOutputFile    = flag.String("o", flagutil.EnvOrDefault("REQUEST_LOG_FILE", ""), "The request output file")
	)
	flag.Var(&flMaxOutputSize, "max-output-size", "Max size of the output file")

	flag.Parse()

	if *flListenAddr == "" {
		log.Fatal("please provide the listen address with -l")
	}

	var output io.Writer
	switch {
	case *flOutputFile != "":
		log.Printf("output file: %s", *flOutputFile)

		tmp, err := os.OpenFile(*flOutputFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600)
		if err != nil {
			log.Fatalf("unable to open file %s. err=%v", *flOutputFile, err)
		}
		output = &checkSizeFileWriter{
			maxSize: int64(flMaxOutputSize),
			f:       tmp,
		}

	default:
		output = ioutil.Discard
	}

	handler := &dumper{
		logFn:  log.Printf,
		output: output,
	}
	mux := http.NewServeMux()
	mux.Handle("/", handler)

	server := &http.Server{
		Addr:    *flListenAddr,
		Handler: mux,
	}
	log.Printf("listening on %s", *flListenAddr)
	go server.ListenAndServe()

	// wait indefinitely

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)

	<-ch

	// shutdown

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("unable to shutdown server properly. err=%v", err)
	}
}
