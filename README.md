# http_status #
#### a tiny HTTP health status service ####

http_status是一个精简的httpd服务器，能够执行简单的类似status.html静态文件的作用，使用方法如下：

	GET /status?PORT=80&SERVICE=squid&VIP=1.2.3.4

服务器使用方法

	Usage: http_status [options]
	
	Options:
	  -h, --help            show this help message and exit
	  -p PORT, --port=PORT  port to listen
	  -i IP, --ip=IP        ip address to bind


服务器将会在/var/run/http_status/目录下检测相应的service文件，如存在，则返回200，否则返回404。如定义VIP也会增加探测VIP地址。

URL处理方式：

1. 只支持GET协议，其他不支持
2. 只定义了/status接口(目录），其他不支持
3. 定义了SERVICE、PORT、VIP 3个变量
   * SERVICE 如不定义则默认文件为status.html
   * PORT 定义后，则检测文件为SERVICE.PORT，如上例中则为squid.80
   * VIP 地址定义后，将会检测是否有VIP的loopback地址

http_status.functions设计用来给init脚本用来直接引用的，如果init脚本中source了http_status.functions，那么启动脚本将会有两个指令：

* "--activate" 在/var/run/http_status下建立以这个服务为名字的空文件，服务上线
* "--inactivate" 在/var/run/http_status下删除以这个服务为名字的空文件，服务下线

这样会更统一，更方便使用：

	/etc/init.d/trafficserver activate
	Activating trafficserver :                                 [  OK  ]

*注意，上线下线的生效时间需要根据检测端的探测间隔来计算*

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


## 增强计划与问题
1. daemon的模式比较渣，需要后续改进
2. 其他init系统如upstart、systemd等是否也支持？
3. 我们需要记录日志吗？
4. 性能是否是问题？
5. 如何提高系统安全性？
6. TODO: should we make statusdir a runtime config?