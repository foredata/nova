# nova[WIP]

## 简介

**nova**是一套完整的微服务开发框架(尚在完善中),它的很多模块都参考了[go-micro](https://github.com/asim/go-micro)，但实现上又有所不同,这里更强调高性能，易扩展，易集成

## 设计准则

- 高性能
    - 类似netty,异步网络接口，提供gpc(goroutine per connection)和nio两种传输模型,基于链表的buffer管理
- 易扩展
    - 基于接口的设计方便扩展
- 易上手
    - 每个模块都会有一个默认实现，便于开发调试，但并非最优实现
- 零依赖
    - 原则上尽量只依赖标准库,对于一些小的第三方库,直接集成在代码中,比较庞大的第三方库需要放到plugin中实现并在使用时手动注册

## 核心模块

- **network**
- **database**
    - redis
    - sqlx
- **store**
- **job**
- **cache**
- **idgen**
- **timing**
- **locking**
- **ddd(Domain-Driven Design)**
- **transaction**
- **feature**
- **debug**
    - logs
    - metrics
    - tracing

## 使用场景

TODO

## 示例代码

TODO

## 集成或参考的第三方库

- [go-micro](https://github.com/micro/go-micro)
- [backoff](https://github.com/cenkalti/backoff)
- [backoff](https://github.com/rfyiamcool/backoff)
- [shortid](https://github.com/teris-io/shortid)
- [xid](https://github.com/rs/xid)
- [hashstructure](https://github.com/mitchellh/hashstructure)
- [mergo](https://github.com/imdario/mergo)
- [base58](https://github.com/mr-tron/base58)
- [fsnotify](https://github.com/fsnotify/fsnotify)
- [dateparse](https://github.com/araddon/dateparse)