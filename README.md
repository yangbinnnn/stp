## 基本使用
隧道建立流程:
1. stpcli 连接stpsrv 控制端口(websocket)
2. stpcli 登录成功，stpsrv 返回分配的端口，连接配置信息和stpsrv　服务端SSH　私钥
3. stpcli 通过分配的端口和私钥与远程建立隧道，默认将本地22 端口映射到远程分配的端口

### 服务端
- 创建用户并创建id_rsa, 安全起见最好最好不要用root

```
> useradd tunnel
> passwd tunnel
> su tunnel
> ssh-keygen
> cat id_rsa.pub > authorized_keys
> chmod 0600 authorized_keys
```

- 配置cfg.json

```
{
    "authKey": "tunnelkey",
    "PortRange": "10001-20000",
    "listenAddr": ":10000",
    "sshAddr": "127.0.0.1:22",
    "sshUser": "tunnel",
    "sshRsaPath": ""
}

```

- 开启服务, 安全起见默认用户为op, `-d` 后台运行

```
>  ./stpsrv -d
```

### 客户端
- 连接服务端并建立隧道

```
> ./stpcli -n yangbin
2018/05/04 17:20:00 ssh user: tunnel
2018/05/04 17:20:00 ssh addr: 127.0.0.1:22
2018/05/04 17:20:00 assgin port: 12345
```

- 后台服务运行`-d`

```
> ./stpcli -n yangbin -key tunnelkey -d
```

### 管理客户端
- 枚举在线客户端

```
[tunnel@op yangbin]$ ./stpsrv -l
+-----+---------+-------+-------------------+--------+
| NUM |  NAME   | PORT  |       ADDR        | ONLINE |
+-----+---------+-------+-------------------+--------+
|   0 | yangbin | 12345 | 172.17.17.4:36648 | true   |
+-----+---------+-------+-------------------+--------+
```

- 指定客户端编号进入远端服务器

```
> [tunnel@op yangbin]$ ./stpsrv -c 0
root@127.0.0.1's password: 
```


### 其他

- 某些情况下需要映射web 服务端口等，可以用`-p` 只能一个端口，如80

```
$ ./stpcli -n yangbin -p 80
2018/05/04 17:26:23 ssh user: tunnel
2018/05/04 17:26:23 ssh addr: 127.0.0.1:22
2018/05/04 17:26:23 assgin port: 16649
```

- 在远程服务端访问16649 端口即可

```
[tunnel@op yangbin]$ curl -i http://127.0.0.1:16649
HTTP/1.1 200 OK
Server: openresty/1.13.6.1
Date: Fri, 04 May 2018 09:26:53 GMT
Content-Type: text/html
Content-Length: 562
Last-Modified: Mon, 13 Nov 2017 08:05:52 GMT
Connection: keep-alive
ETag: "5a095260-232"
Accept-Ranges: bytes
```

## 功能清单
- 自动分配隧道端口
- 断线重连
- 离线客户端tunnel清理
- 自动添加publicKey，免密登录
- 后台服务运行
- 优化代码，完善控制逻辑
