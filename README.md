# clicli-cdn
> 适用于 c 站的分块上传 && 视频编解码

### 部署

一、手动部署
0. 准备一台空的 linux 机器，在 root 目录新建文件夹 video

1. 将 bin/cdn 文件上传到 video 文件夹，并设置权限 777

```
nohup ./cdn &
```
2. logo.png 同样需要上传到 video 文件夹，这是视频编码时用到的水印

3. nginx 配置当前目录为静态文件目录

```nginx
     location /video/ {   
        alias /root/video/;     
        index index.html;   
    } 
    location / {                      
        proxy_pass http://localhost:2333;   
    } 
```

