package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	dbConn := &DbConn{}

	jobs := make(chan any)
	workerDone := make(chan struct{})
	worker := Worker{
		workerPoolSize: 10,
		jobs:           jobs,
		done:           workerDone,
		db:             dbConn,
	}

	server := NewServer(8080, 10*time.Second, worker)
	var wg sync.WaitGroup
	wg.Go(func() {
		fmt.Printf("starting server!")
		if err := server.Start(); err != nil {
			fmt.Printf("server start failed: %v", err)
		}
	})

	cacheContext, cacheCancel := context.WithCancel(ctx)
	wg.Go(func() {
		Cache(cacheContext)
	})

	<-done
	fmt.Println("os signal triggered")
	server.Stop(ctx)
	workerDone <- struct{}{}
	cacheCancel()
	dbConn.Close()

	wg.Wait()
	fmt.Println("all gouroutines finished")
}

type Server struct {
	Port           int
	RequestTimeout time.Duration
	Worker         Worker
	httpServer     *http.Server
}

func NewServer(port int, requestTimeout time.Duration, worker Worker) *Server {
	return &Server{
		Port:           port,
		RequestTimeout: requestTimeout,
		Worker:         worker,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /proc", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			time.Sleep(1 * time.Second)
		}()
		w.Write([]byte("processing"))
	})

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Port),
		ReadTimeout:  s.RequestTimeout,
		WriteTimeout: s.RequestTimeout,
		IdleTimeout:  s.RequestTimeout,
		Handler:      mux,
	}
	s.httpServer = httpServer

	return httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

type Worker struct {
	workerPoolSize int
	jobs           chan any
	done           chan struct{}
	db             net.Conn
}

func (w *Worker) Do() { // 10
	for i := 1; i <= w.workerPoolSize; i++ {
		go worker(w.jobs, nil, w.done)
	}
}

func worker(jobs chan any, results chan any, done chan struct{}) {
	select {
	case <-done:
		return
	default:
		for job := range jobs {
			time.Sleep(1 * time.Second)
			fmt.Printf("doing job: %v\n", job)
		}
	}
}

func Cache(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ticker.C:
		fmt.Printf("caching...\n")
	}

	return nil
}

type DbConn struct{}

// Close implements [net.Conn].
func (d *DbConn) Close() error {
	fmt.Println("closing DB Connection")
	return nil
}

// LocalAddr implements [net.Conn].
func (d *DbConn) LocalAddr() net.Addr {
	panic("unimplemented")
}

// Read implements [net.Conn].
func (d *DbConn) Read(b []byte) (n int, err error) {
	panic("unimplemented")
}

// RemoteAddr implements [net.Conn].
func (d *DbConn) RemoteAddr() net.Addr {
	panic("unimplemented")
}

// SetDeadline implements [net.Conn].
func (d *DbConn) SetDeadline(t time.Time) error {
	panic("unimplemented")
}

// SetReadDeadline implements [net.Conn].
func (d *DbConn) SetReadDeadline(t time.Time) error {
	panic("unimplemented")
}

// SetWriteDeadline implements [net.Conn].
func (d *DbConn) SetWriteDeadline(t time.Time) error {
	panic("unimplemented")
}

// Write implements [net.Conn].
func (d *DbConn) Write(b []byte) (n int, err error) {
	fmt.Printf("writing to DB: %v", b)
	return 0, nil
}

var _ net.Conn = (*DbConn)(nil)
