# http_status #
#### a tiny HTTP health status service ####

TODO: add more description on why this service is very useful
  

## HAProxy health check ATS
haproxy里backend里加：

	option httpchk GET /status?SERVICE=trafficserver HTTP/1.1\r\nHost:\ cdn.hc.org
	http-check disable-on-404

ATS remap里加：

	map http://cdn.hc.org:8080/ http://127.0.0.1:8001/
	map http://cdn.hc.org:80/ http://127.0.0.1:8001/

## HAProxy health check Squid

haproxy里backend里加：

	option httpchk GET /status?SERVICE=squid HTTP/1.1\r\nHost:\ cdn.hc.org
	http-check disable-on-404

Squid的peer中添加：

	##squid 7 layer
	cache_peer 127.0.0.1 parent 8001 0 originserver proxy-only
	cache_peer_domain 127.0.0.1  cdn.hc.org

## LVS (keepalived) health check HAProxy
KeepAlived real_server add:

    inhibit_on_failure
    HTTP_GET {
        url {
            path /status?SERVICE=haproxy&VIP=1.2.3.4,1.2.3.5
            status_code 200
        }
        connect_timeout 3
        nb_get_retry 2
        delay_before_retry 5
    }

HAProxy加backend：

	# http_status backend, to localhost
	backend http_status
	  	server hc 127.0.0.1:8001 maxconn 100
  	
HAProxy的frontend加：

	acl hc hdr_beg(host) -i cdn.hc.org
	use_backend http_status if hc