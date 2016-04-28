//http_status is a tiny http health status server
//
// TODO: should we make statusdir a runtime config?
// TODO: should we log access logs?

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
)

const statusdir = "/var/run/http_status"

// Detach from controling terminal and run in the background as system daemon.
// redirects os.Stdin, os.Stdout and os.Stderr to "/dev/null"
func Daemonize() (*os.Process, error) {
	daemonizeState := os.Getenv("_GOLANG_DAEMONIZE_FLAG")
	switch daemonizeState {
	case "":
		syscall.Umask(0)
		os.Setenv("_GOLANG_DAEMONIZE_FLAG", "1")
	case "1":
		syscall.Setsid()
		os.Setenv("_GOLANG_DAEMONIZE_FLAG", "2")
	case "2":
		return nil, nil
	}

	var attrs os.ProcAttr

	attrs.Dir = "/tmp"
	f, err := os.Open("/dev/null")
	if err != nil {
		return nil, err
	}
	attrs.Files = []*os.File{f, f, f}

	exe, err := filepath.Abs(os.Args[0])
	if err != nil {
		return nil, err
	}

	realexe, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return nil, err
	}

	p, err := os.StartProcess(realexe, os.Args, &attrs)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func main() {

	// we have only two option flags to adapt, listen ip address and port.
	var flagport uint
	var flagip string
	var godaemon bool
	var pidfile string
	flag.UintVar(&flagport, "port", 8001, "port to listen")
	flag.UintVar(&flagport, "p", 8001, "port to listen, in short")
	flag.BoolVar(&godaemon, "daemon", false, "go daemon")
	flag.BoolVar(&godaemon, "d", false, "go daemon")
	flag.StringVar(&pidfile, "pidfile", "", "ip address to bind")
	flag.StringVar(&flagip, "ip", "127.0.0.1", "ip address to bind")
	flag.StringVar(&flagip, "i", "127.0.0.1", "ip address to bind, in short")
	flag.Parse()

	listenaddr := fmt.Sprint(flagip, ":", flagport)

	// fmt.Printf("Starting http_status on: %s\n", listenaddr)

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		// Setup Server response env
		w.Header().Set("Server", "Status")
		w.Header().Set("Cache-Control", "private")

		requesturl := r.URL.Query()

		var statusfile string = fmt.Sprint(statusdir, "/", "status.html")
		// if we have a SERVICE= in Query, then we should change the statusfile
		service, ok := requesturl["SERVICE"]
		if ok {
			statusfile = fmt.Sprint(statusdir, "/", service[0])
		}
		// if we have a PORT= in Query, then we should append .PORT in the statusfile
		port, ok := requesturl["PORT"]
		if ok {
			statusfile = fmt.Sprint(statusfile, ".", port[0])
		}

		// well, let us check if the statusfile exist,
		// and check the VIP if needed
		_, statusfileresult := os.Stat(statusfile)
		if os.IsNotExist(statusfileresult) {
			http.Error(w, "Not Found - Service not aviliable!", http.StatusNotFound)
		} else {
			// VIP checking
			vip, ok := requesturl["VIP"]
			if ok {
				vip[0] = fmt.Sprint(vip[0], "/32")

				var vipfine bool = false

				// the VIP should be a /32 masked alias on loop device
				nics, _ := net.Interfaces()
				for _, nic := range nics {
					if nic.Flags&net.FlagLoopback == net.FlagLoopback {
						loopaddrs, _ := nic.Addrs()
						for _, addr := range loopaddrs {
							if vip[0] == addr.String() {
								// VIP check passed, let us return 200
								vipfine = true
								http.Error(w, "OK - service and VIP all fine!", http.StatusOK)
							}
						}
					}
				}
				// VIP check fail?
				if !vipfine {
					http.Error(w, "Not Found - VIP alias not aviliable!", http.StatusNotFound)
				}
			} else {
				// no VIP check needed, fine
				http.Error(w, "OK - service fine!", http.StatusOK)
			}
			// end VIP checking
		}

	})

	if godaemon {
		p, err := Daemonize()
		if err != nil {
			log.Fatalf("Daemonize failed: %v", err)
		}

		if p != nil {
			// parent process, nop
			return
		} else {
			// let's do the work

			if pidfile != "" {
				// we don't guarantee that pidfile is not created already
				// TODO: add pidfile locking to prevent run, do we need?
				abspidfile, _ := filepath.Abs(pidfile)
				pidf, err := os.OpenFile(abspidfile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
				if err == nil {
					fmt.Fprint(pidf, os.Getpid())
					pidf.Close()
					defer os.Remove(abspidfile) // hmm, can we remove after kill?
				}
			}

			// let's go
			log.Fatal(http.ListenAndServe(listenaddr, nil))
		}

	} else {
		// don't go daemon, run it
		log.Fatal(http.ListenAndServe(listenaddr, nil))
	}
}
