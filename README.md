![license](https://img.shields.io/github/license/919927181/rdr)
![download](https://img.shields.io/github/downloads/919927181/rdr/total)
[![Build Status](https://github.com/919927181/rdr/actions/workflows/go.yml/badge.svg)](https://github.com/919927181/rdr/actions?query=branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/919927181/rdr)](https://goreportcard.com/report/github.com/919927181/rdr)


RDR: redis data reveal
=================================================

## About（介绍）

RDR (redis data Reveal) is a tool for offline analysis of redis rdb files. Through it, you can quickly discover bigkeys, help you grasp the occupation and distribution of keys in memory, learn which keys are growing infinitely (through key expiration time or quantity). It provides data support for your optimization operations and helps you avoid problems such as insufficient memory and performance degradation caused by key skew.

RDR(redis data Reveal)是一个用于离线分析 redis rdb 文件的工具。通过它，可以快速发现实例中的bigkey，帮助您掌握Key在内存中的占用和分布情况、得知哪些key在无限增长等。能为您的优化操作提供数据支持，帮助您避免因Key倾斜（导致集群内存分布不均）引发的内存不足、性能下降等问题。

 功能：
  - 统计信息展示（command:show）：以网页形式展示RDB文件的统计报告（例如，按数据类型分布、Top 300 大Key列表和按前缀分析等）。
  - 统计信息保存(command:dump2file)：除了在线网页展示外，还可以将统计信息保存到文件。
  - 获取所有key（command:keys）：从RDB文件中解析出全部键名以及属性信息（数据类型、内存大小、元素数量、过期时间、所属db等），保存到文件，以便自行分析。
 
 备注：如果你想知道集合类中的哪个元素最大，请将统计信息导出到文件，见报告中的LargestKeys>FieldOfLargestElem字段。
 
 特点：
  - 安全无扰：分析过程完全在 RDB 备份文件上进行，对线上Redis实例零影响。
  - 使用方便：提供了linux和windows下的可执行文件，不需要安装；一键生成内存健康报告，在线图形化展示更直观。
  - 高效解析：RDR由golang实现，解析速度比较快，一个10G的rdb文件，用时不到4分钟。
  - 结果精准：结果反映的是RDB快照生成时刻的精确状态。
  - 庖丁解牛：深入RDB文件结构与LRU元数据原理，犹如为缓存做了一次精准的“核磁共振”检查。

注意：生产环境下慎行，以免对线上实例产生性能影响，我们可以在从节点或拷贝rdb文件到测试机上去执行。

## Fork（基于）

This repository is fork  from github.com/xueqiu/rdr.  The requrie rdb file parse is github.com/dongmx/rdb ，in this project has been replaced  github.com/919927181/rdb

- 本项目基于 xueqiu/rdr 开源项目开发，实现了对redis7+的支持。
- 核心依赖包 919927181/rdb（源自dongmx/rdb）解析redis rdb 文件。

注：xueqiu/rdr是雪球公司基于redis-rdb-tool开源项目开发的，更新维护停止在 2019 年 10 月 9 日。

在此，对原开源项目作者，以及提供issues、pr的朋友们，表示感谢。


## Support For Redis（redis 版本支持）

支持redis rdb 文件版本为 1 <= version <= 12

  - RDR V1.0.3 支持 Redis6（redis 5.x ~ 6.x ，rdb的版本是 9 ) 
  - RDR V1.0.5 支持 Redis7.0+（rdb文件版本10~12，mysql8.0的rdb版本是12）
  
 备注：
  - 针对redis7+版本，rdb文件解析主要是解决listpack数据类型问题。
  - RDR的核心依赖是 rdb 文件解析，不同版本的 redis，其 rdb 文件存在差异，也会增加新的数据类型，存在数据兼容性问题。
    - 如果解析高版本redis时出现错误，可以尝试通过 RedisShake 数据迁移工具，例如将redis7 RDB数据迁移到redis6下，然后再用rdr进行分析。
	
 注意：虽然我们通常不用redis消息队列功能，但是考虑到有人用到了，因此在v1.1.2版本进行了支持，注意，我没做对比验证。

## Change（变更）
- caiqing0204：增加了key所属DB，这样可以更直观的查看key元信息。
- 泰山李工（我）：
   - v1.0.2
     - 将依赖 github.com/dongmx/rdb 中的rdbVersion 由9改成20【2025-11-08】
     - 修改html布局、将标题英文改为中文 【2025-11-08】
	 
   - v1.0.3 
     - 升级chartjs版本，实现图表tip时，显示更人性化的数字【2025-11-13】
     - 将2021年3月 至 2023年7月，在原项目 [xueqiu/rdr](github.com/xueqiu/rdr/pulls) 的pulls，均已同步过来、并解决完毕。
	 
   - v1.0.5 
     - 完成redis7+支持，主要解决了redis7.x底层存储类型使用listpack替代ziplist的解析问题。


## Usage（使用）

```
USAGE:
   rdr [global options] command [command options] [arguments...]

VERSION:
   vx.x.x

COMMANDS:
     dump2file dump statistical information of rdbfile to file(./rdb-report-xxx.json).
     show      show statistical information of rdbfile by webpage
     keys      get all keys from rdbfile, write to file（/tmp/rdb-all-keys-xxx.txt）.
     help, h   show a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

```
NAME:
   rdr show - show statistical information of rdbfile by webpage

USAGE:
   rdr show [command options] FILE1 [FILE2] [FILE3]...

OPTIONS:
   --port value, -p value  Port for rdr to listen (default: 8080)
```


### linux下使用说明

```
1.从releases中，下载 linux 下的可执行文件

2.创建目录
# mkdir -p /tmp/rdb/
# cd /tmp/rdb/

3.然后把rdr工具、redis的数据库文件.rdb上传到该目录下

给工具赋予执行权限
# chmod a+x ./rdr*

4.运行
# GOGC=200 ./rdr-linux show -p 8099 *.rdb
注意,如果你的rdb文件比较大（1G+）,建议一次只分析一个rdb文件
    如果rdb文件大，那么cpu使用率就会过高，此时我们调整GOGC，默认100，提高值(200-400)可降低GC频率，减少CPU占用但会增加内存使用

5.防火墙端口放行
     For Ubuntu\Debian：sudo ufw allow 8099/tcp  &&  sudo ufw reload
     For Redhat\Centos：
          sudo firewall-cmd --zone=public --add-port=8099/tcp --permanent
           sudo firewall-cmd --reload
		   
6.查看分析报告，浏览器访问 http://your-host:8099/
```

### windows下使用说明
```
 1. 从releases中，下载 windows 下的可执行文件

 2. 将rdr工具、redis的数据库文件.rdb拷贝到某个文件夹下

 3. 在该文件夹下空白的地方，ctrl+鼠标右击，打开 cmd 执行：
    > .\rdr-win64.exe show -p 8099 dump.rdb

 4. 查看分析报告，打开浏览器访问 http://localhost:8099 
```

## Exapmle 使用示例
```
# 1.通过网页显示rdb file的统计报告
$ GOGC=200 ./rdr-linux show -p 8080 dump.rdb
```
Note that the memory usage is approximate.
<img width="1155" height="612" alt="image" src="https://github.com/user-attachments/assets/a8b16a78-b232-4282-b2ff-781f0cc87504" />

```
# 2.将统计报告信息写到文件（当前目录/rdb-report-xxx.json）
# 如果你想知道集合类中的哪个元素最大，见报告中的LargestKeys>FieldOfLargestElem字段。
$ GOGC=200 ./rdr-linux  dump2file  dump.rdb
# 获取 RDB 文件中最大的 N 个键时，N默认为100，最大500; 加上内存大小过滤参数-s, 就可以过滤掉小于该阈值的key，支持的单位为B/KB/MB/GB
$ GOGC=200 ./rdr-linux  dump2file -n 300 -s 10kb dump.rdb

# 解析出所有key及属性信息，输出到文件（当前目录/rdb-all-keys-xxx.txt），以便自行分析之需要。
$ GOGC=200 ./rdr-linux keys dump.rdb
key,type,encoding,size,humanizeSize,numOfElem,expiration,lruIdle,lfuFreq,db
student:1:name, string, string, 100, 100 B, 8, , , 0, 0
colors, set, listpack, 94, 94 B, 2, , , 0, 0

附-mysql建表语句：
CREATE TABLE rdb_keys_infor (
   `id`  int  PRIMARY KEY AUTO_INCREMENT,
    `key` varchar(300)  NOT NULL COMMENT 'key名',
    `type` varchar(30)  NOT NULL COMMENT '类型',
    `encoding` varchar(30)  NOT NULL COMMENT '底层类型',
    `size` int NOT NULL COMMENT '内存使用',
    `humanizeSize` varchar(30) NOT NULL COMMENT '内存使用',
    `numOfElem` int NOT NULL COMMENT '元素数量',
    `expiration`  varchar(30)   COMMENT '过期时间',
    `lruIdle`  varchar(30)   COMMENT '最后一次方式时间',
    `lfuFreq` int NOT NULL COMMENT '访问频率',
    `db` int NOT NULL COMMENT 'db'
) ENGINE = InnoDB COMMENT = 'redis key信息表';

备注，根据maxmemory-policy配置的淘汰策略，我们可以找出冷数据：
  - 如果策略是volatile-lru或allkeys-lru，记录的是Key的最后一次访问时间（lruIdle）。
  - 如果策略包含LFU（如volatile-lfu），记录的是访问频率（lfuFreq）。

```

## 常见问题

```
Q：为什么使用 memory usage 命令和rdr算的内存使用不一致？
A：Key和value所对应的struct和指针大小。在jemalloc分配后，字节对齐部分所占用的大小也会计算在used_memory中
   rdr分析的key内存占用是一个近似值。无论是用命令还是rdr都计算了这两块，为什么不一致？可读下 https://blog.csdn.net/f80407515/article/details/122387859

Q：如何处理报错decode rdbfile error: rdb: unknown object type 116 for key？
A：该报错表示实例中存在非标准或新版本增加的数据结构，暂不支持分析，你可以还原到测试实例删除后再进行分析。

Q：为什么Redis缓存分析中String类型Key的元素数量和元素长度是一样的？
A：在Redis缓存分析中，针对String类型的Key，其元素数量就是其元素长度。

Q：Redis缓存分析的前缀分隔符是什么？
A：目前Redis缓存分析的前缀分隔符是按照固定的前缀:;,_-+@=|# 区分的字符串。

Q：各key的内存占用为什么比[HDT3213/rdb](https://github.com/HDT3213/rdb)算的大28？
A：lru_bits 默认占用24比特位，HDT3213/rdb V.1.3.0没有计算，而本工具计算了，请看源码d.m.TopLevelObjOverhead。

Q：通常redis集群只会用到db0,单例中可能会用多个槽。那么当不同db里有相同前缀的key时，前缀分析列表该如何显示所属db？
A: 从v1.0.9版本起，将所有所属的db都进行了显示，多个时以逗号隔开。
```


## RDR 开发

1. 文件目录结构

```
\decoder\
   |-- decoder.go       # 对github.com/919927181/rdb/decoder.go中的interface方法实现，rdb解析各类型数据时，会调用相应的方法，进行计数。
   |-- memprofiler.go  # 内存分析器
\dump\
   |-- dump.go          # 将rdb文件统计信息输出到STDOUT或文件
   |-- keys_export.go  # 将获取所有key及属性信息输出到File，以便自行分析之需要
\static  # 以html展示结果时，需要的静态资源文件
\views  # html 前端页面代码

```

2. 如果你想修改redis rdb 文件解析插件的源码，可以pr到github.com/919927181/rdb

   你可以直接修改下载到本地的依赖 \vendor\github.com\919927181\rdb，调试成功后，再进行pr 或创建\并引用自己的rdb依赖

3. 如果你需要修改html

    你需要安装go-bindata，安装手册可参考 https://blog.csdn.net/qq_67017602/article/details/130742316

4. 打包
   
```
 1. 在windows下打包，编译出 linux 下的可执行文件，在项目根目录下，打开cmd，执行以下命令
    set CGO_ENABLED=0
    set GOOS=linux
    set GOARCH=amd64
    go build -o rdr-linux  main.go
	
	编译出Windows下的exe文件
	set CGO_ENABLED=0
	set GOOS=windows
	set GOARCH=amd64
	go build -o rdr-win64.exe  main.go

 2.如果改动了静态资源（css\js\html），需要使用go-bindata将静态资源文件嵌入到go文件里
    go-bindata -prefix "static/" -o=static/static.go -pkg=static -ignore static.go static/... 
    go-bindata -prefix "views/" -o=views/views.go -pkg=views -ignore views.go views/...
 
 3. 在编译前自动化生成某类代码，它常用于自动生成代码，我一般是直接执行打包命令
    go generate
```


## RDB 开发

 rdr工具的核心部分就是rdb文件解析，作为开发者，我们可以通过以下几个途径来掌握相关知识：

1. 大部分 rdb 文件的解析都是按照 https://github.com/sripathikrishnan/redis-rdb-tools/wiki/Redis-RDB-Dump-File-Format  和 github.com/cupcake/rdb 来的，
   RDB 文件格式说明：https://www.cnblogs.com/Finley/p/16251360.html ，完整解析器源码在 github.com/HDT3213/rdb

2. Redis迁移工具RedisShake（语言golang），从 RDB 文件中读取数据写入目标端，我们可以参考这块的代码。
   项目地址：https://github.com/tair-opensource/RedisShake
   rdb.go源码地址：https://github.com/tair-opensource/RedisShake/tree/v4/internal/rdb

3. 对照 redis 源码

```
  rdb.c 文件：https://github.com/redis/redis/blob/7.0-rc3/src/rdb.c     // RDB 文件读写，行1736：robj *rdbLoadObject
  rdb.h 文件：https://github.com/redis/redis/blob/unstable/src/rdb.h   // RDB version、RDB object types 等定义

```

注：  
   - rdb 对数字这一块的解码操作要特别注意，不一定能用 BitConverter.ToIntXX 来获得正确的值！！ 
   - redis7.x底层存储类型使用listpack替代ziplist。例如，若List大小超过阈值（list-max-listpack-size），Redis会切换为ziplist或quicklist编码


## 贡献

欢迎来自各界的贡献。对于重大变更，请先开一个 issue 来讨论你想要改变的内容。如果想共同维护此项目，请加我微信（Sd-LiYanJing）。

特别感兴趣的是：

 1. 随着redis版本变化，增加新类型的解析支持
 2. 优化、改善代码，提升性能
 

## 交流群

添加微信（Sd-LiYanJing），备注GitHub-rdr，即可进群

最后，欢迎Star，欢迎开发者加入！


## License

This project is under Apache v2 License. See the [LICENSE](LICENSE) file for the full license text.