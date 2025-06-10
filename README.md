# Introduction
## 项目背景
echat是一个基于go语言的工单系统客服系统，主要功能有：
- 基于工单系统创建聊天室进行聊天
- 发送图片、文字

# Prework
## 1. 创建docker network, docker volume [ production mode ]
```
docker network create echat-network
docker volume create echat-postgres-data
```
## 2. 安装postgres，初始化数据库
```
# install postgres
docker run -dit --name postgres --network echat-network -v echat-postgres-data:/var/lib/postgresql/data -e POSTGRES_PASSWORD=$DB_PASSWORD -p 5432:5432 docker.m.daocloud.io/postgres:13-alpine3.22

# create database
docker exec -it postgres psql -U postgres -c "CREATE DATABASE echat;"
```

# 运行
## Development mode
### 1. 准备env文件
### 2. 运行
```
source .env
go generate
go run main.go
```
## Production mode
### 1. build docker image
```
docker build -t echat-server:latest .
```
### 2. 运行容器
#### 1. create docker volume
```
# 创建docker volume
docker volume create echat-server-uploads
```
#### 2. set env  
```
# 设置env
export DB_PASSWORD=<数据库密码>
export DB_NAME=<数据库名>
export ITSM_SERVER_URL=<itsm服务地址>
```
```
# example
export DB_PASSWORD='password#123'
export DB_NAME='echat'
export ITSM_SERVER_URL='http://10.8.7.171:9999/api/itsmServer/getInstance/detail/listbyhandler'
```


#### 3. 运行容器
```
docker run -dit --name echat-server \
    --network echat-network \
    -v echat-server-uploads:/echat/public \
    -e DB_HOST='postgres' \
    -e DB_PORT='5432' \
    -e DB_USERNAME='postgres' \
    -e DB_PASSWORD=$DB_PASSWORD \
    -e DB_NAME=$DB_NAME  \
    -e ITSM_SERVER_URL=$ITSM_SERVER_URL \
    -p 8180:8080 echat-server:latest
```