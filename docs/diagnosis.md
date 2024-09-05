# diagnosis (诊断)

diagnosis 是一个C/S模式的诊断工具，包含诊断端和数据接收端。诊断端通过webSocket连上服务端，接受服务端的控制。

计划规划到camp系统中。



app: 诊断客户端App




## 用户诊断工具

ICMP

DNS

HTTP & HTTPS

MTR / traceroute

TELNET / Socket Conn



## OneClickDiag

客户端通过一键诊断工具可以快速将需要的信息进行收集，然后上报到服务端，方便服务端人员进行定位诊断。


上报字段：

    requestId：全局唯一 uuid，通过该 uuid 可以快速查找用户上报信息。

    clientIP：客户端公网 IP 地址
    
    BasicNetWorkBody: 基础网络相关字段
        LocalDNS：
        MTR：
    
    AppNetWorkBody：应用网络相关字段
        域名解析结果
    
    AppBody：某项上传内容
        应用访问结果
    










