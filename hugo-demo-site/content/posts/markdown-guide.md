---
title: "Markdown 语法速览"
date: 2026-08-02T10:30:00+08:00
draft: false
tags: ["Markdown", "写作"]
categories: ["教程"]
description: "在 Hugo 中常用的 Markdown 语法示例"
---

Hugo 默认使用 Goldmark 作为 Markdown 渲染引擎，支持标准 Markdown 语法以及部分扩展。

## 标题

用 `#` 的数量表示标题级别，从一级到六级。

## 列表

无序列表：

- 苹果
- 香蕉
- 橙子

有序列表：

1. 第一步：安装 Hugo
2. 第二步：创建站点
3. 第三步：编写内容并预览

## 表格

| 命令 | 作用 |
| --- | --- |
| `hugo new site xxx` | 创建新站点 |
| `hugo new posts/xxx.md` | 新建一篇文章 |
| `hugo server -D` | 本地预览（含草稿） |
| `hugo --minify` | 生产构建并压缩输出 |

## 引用

> 静态网站生成器把"内容"和"呈现"彻底解耦，这正是它优雅的地方。

## 代码块

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Hugo!")
}
```

## 链接与强调

这是一个 [Hugo 官网](https://gohugo.io) 的链接。**加粗文本** 和 *斜体文本* 也都支持。
