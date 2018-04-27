package logger

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Logger struct {
	unixSockFile string
	w            io.Writer
	r            io.Reader
}

func loop(ln net.Listener) {

	for {
		fd, err := ln.Accept()
		if err != nil {
			log.Fatal("Accept error: ", err)
		}

		go echoServer(fd)
	}
}

func CreateLogger(sock string) Logger {

	if sock != "" {
		os.Remove(sock)
	}
	log.Println("Starting echo server")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(ln net.Listener, c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		ln.Close()
		os.Exit(0)
	}(ln, sigc)

	go loop(ln)

	c, err := net.Dial("unix", sock)
	if err != nil {
		log.Fatal("Dial error", err)
	}

	//defer c.Close()

	go reader(c)

	return Logger{unixSockFile: sock, w: c, r: c}
}

func echoServer(c net.Conn) {

	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}

		data := buf[0:nr]
		println("Server got:", string(data))
		_, err = c.Write(data)
		if err != nil {
			log.Fatal("Writing client error: ", err)
		}
	}
}

func reader(r io.Reader) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			return
		}
		println("Client got:", string(buf[0:n]))
	}
}

func (lg *Logger) Write(msg string) {

	_, err := lg.w.Write([]byte(msg))
	if err != nil {
		log.Fatal("Write error:", err)

	}
}

func (l *Logger) Writer() io.Writer {
	return l.w
}
