---
title: "欢迎使用 Hugo"
date: 2026-08-01T09:00:00+08:00
draft: false
tags: ["Hugo", "入门"]
categories: ["教程"]
description: "第一篇示例文章，介绍 Hugo 的基本概念"
---

## 什么是 Hugo

Hugo 是一个使用 **Go** 语言编写的开源静态网站生成器，以极致的构建速度著称——即便是数以万计的页面，通常也能在几秒内完成渲染。

## 核心工作流程

Hugo 的基本工作方式可以概括为三步：

1. 在 `content/` 目录中用 Markdown 编写内容
2. 在 `layouts/` 目录中定义模板，决定内容如何被渲染成 HTML
3. 执行 `hugo` 命令，生成的静态文件会输出到 `public/` 目录

```bash
# 本地开发,启动带热重载的预览服务器
hugo server -D

# 生产构建,输出到 public/ 目录
hugo --minify
```

## 为什么选择静态网站生成器

相比传统的动态 CMS（如 WordPress），静态站点在以下方面有明显优势：

- **速度**：没有数据库查询和服务端渲染开销，页面直接是纯 HTML
- **安全**：没有后台管理系统和数据库暴露在外网，攻击面小得多
- **成本**：可以免费部署在 GitHub Pages、Cloudflare Pages、Netlify、Vercel 等平台
- **版本控制友好**：所有内容都是文本文件，天然适合 Git 管理

继续阅读下一篇，了解 Markdown 的常用语法。
