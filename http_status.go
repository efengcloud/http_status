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
)

const statusdir = "/var/run/http_status"

func main() {

	// we have only two option flags to adapt, listen ip address and port.
	var flagport uint
	var flagip string
	flag.UintVar(&flagport, "port", 8001, "port to listen")
	flag.UintVar(&flagport, "p", 8001, "port to listen, in short")
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

	log.Fatal(http.ListenAndServe(listenaddr, nil))
}
