# metrics

## 使用方法

## 类型

- Counter: 计数器，单调递增或递减
- Gauge: 计量器，统计瞬时状态的数据信息。
- Meters: 度量某个时间段的平均处理次数（request  per second）
- Histogram: 直方图,统计数据的分布情况，最大值、最小值、平均值、中位数，百分比（75%、90%、95%、98%、99%和99.9%）
- Summary: .
- Timers: 计时器, 统计某一块代码段的执行时间以及其分布情况，基于Histograms和Meters来实现的。

## libraries

- [Metrics介绍和Spring的集成](https://colobu.com/2014/08/08/Metrics-and-Spring-Integration/)
- [Java Metrics](http://t.zoukankan.com/hrhguanli-p-4051995.html)
- [writing_clientlibs](https://prometheus.io/docs/instrumenting/writing_clientlibs/)
- [prometheus client](https://github.com/prometheus/client_golang)
- [influxdb client](https://github.com/influxdata/influxdb-client-go)
- [tally](https://github.com/uber-go/tally)
- [go-metrics](https://github.com/rcrowley/go-metrics) 
- [golang metrics 的基本使用](https://zhuanlan.zhihu.com/p/30441529)
- [datadog metrics type](https://docs.datadoghq.com/metrics/types/?tab=count)
- [meter实现](http://vearne.cc/archives/421)
    - EWMA(Exponentially Weighted Moving-Average) 中文译为指数加权移动平均法
- [dropwizard](https://metrics.dropwizard.io/3.1.0/manual/core/)

- robustperception 讲的比较清晰
    - [how-does-a-prometheus-counter-work](https://www.robustperception.io/how-does-a-prometheus-counter-work)
    - [how-does-a-prometheus-gauge-work](https://www.robustperception.io/how-does-a-prometheus-gauge-work)
    - [how-does-a-prometheus-summary-work](https://www.robustperception.io/how-does-a-prometheus-summary-work)
    - [how-does-a-prometheus-histogram-work](https://www.robustperception.io/how-does-a-prometheus-histogram-work)