//http_status 是一个用来进行为squid/haproxy/lvs等进行7层健康检测而设计的
//精简的httpd服务器，能够执行简单的类似status.html静态文件的作用，使用方法
//如下：
//GET /status?PORT=80&SERVICE=squid&VIP=1.2.3.4
//服务器使用方法
//Usage: http_status [options]
//
//Options:
//  -h, --help            show this help message and exit
//  -p PORT, --port=PORT  port to listen
//  -i IP, --ip=IP        ip address to bind
//服务器将会在/var/run/http_status/检测相应的文件，如存在，则返回200，否则返回
//404。如定义VIP也会增加探测VIP地址。
//URL处理方式：
//1，只支持GET协议，其他不支持
//2，只定义了/status接口(目录），其他不支持
//3，定义了SERVICE、PORT、VIP 3个变量
//   SERVICE如不定义则默认文件为status.html
//   PORT定义后，则检测文件为SERVICE.PORT，如上例中则为squid.80
//   VIP地址定义后，将会检测是否有VIP的loopback地址
//
//配合启动脚本的"--activate" "--inactivate"会更方便使用：
// /etc/init.d/trafficserver activate
//to use the (in)activate, your init script should:
//
//
// TODO: should we make statusdir a runtime config?

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
