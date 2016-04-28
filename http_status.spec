Summary:	http_status - L7 http health check server
Summary(zh_CN.UTF-8):	http_status - 7层的健康检测服务器
Name:		http_status
Version:	0.0.6
Release:	%_release
Vendor:		Zhao Yongming
Packager: 	Zhao Yongming <zym@efengcloud.com>
License:	Apache 2
Group:		Ops/CDN
Source0:	http_status.go
Source1:	http_status.init
Source2:	http_status.functions
URL:		http://www.efengcloud.com
BuildRequires:	golang
BuildArchitectures: x86_64

%description -l zh_CN.UTF-8
http_status 是一个用来进行为squid/haproxy/lvs等进行7层健康检测而设计的
精简的httpd服务器，能够执行简单的类似status.html静态文件的作用，使用方法
如下：
GET /status?PORT=80&SERVICE=squid&VIP=1.2.3.4
服务器使用方法
Usage: http_status [options]

Options:
  -h, --help            show this help message and exit
  -p PORT, --port=PORT  port to listen
  -i IP, --ip=IP        ip address to bind
服务器将会在/var/run/http_status/检测相应的文件，如存在，则返回200，否则返回
404。如定义VIP也会增加探测VIP地址。
URL处理方式：
1，只支持GET协议，其他不支持
2，只定义了/status接口，其他不支持
3，定义了SERVICE、PORT、VIP 3个变量
   SERVICE如不定义则为status
   PORT定义后，则检测文件为SERVICE.PORT，如上例中则为squid.80
   VIP地址定义后，将会检测是否有VIP的loopback地址。

%description
http_status is a simple http server for L7 heath check, it can works for squid
haproxy lvs etc. As a standalone simple httpd server, it is very limited.

%prep

%build
go build

%install
install -m 755 -D %SOURCE1 $RPM_BUILD_ROOT%{_sysconfdir}/init.d/http_status
install -m 644 -D %SOURCE2 $RPM_BUILD_ROOT%{_sysconfdir}/init.d/http_status.functions
install -m 755 -D http_status $RPM_BUILD_ROOT%{_bindir}/http_status
install -d $RPM_BUILD_ROOT/var/run/http_status
%clean
#rm -rf $RPM_BUILD_ROOT

%post
/sbin/chkconfig --add http_status

%postun

%files
%defattr(644,root,root,755)
%attr(0755,root,root) %{_bindir}/http_status
%attr(0755,root,root) %{_sysconfdir}/init.d/http_status
%attr(0755,root,root) %{_sysconfdir}/init.d/http_status.functions
/var/run/http_status

#%define date	%(echo `LC_ALL="C" date +"%a %b %d %Y"`)
%changelog
* Mon Sep 28 2015 Zhao Yongming <zym@efengcloud.com> - 0.0.1
- init commit with golang language.
- more docs to go
