# Hugo 完整示例站点

一个使用 [Hugo](https://gohugo.io)（Go 编写的静态网站生成器）搭建的完整可运行示例工程。主题为自定义手写模板，未依赖任何第三方 theme，方便学习 Hugo 模板系统的每一层。

## 目录结构

```
hugo-demo-site/
├── archetypes/           # hugo new 命令使用的内容模板
│   └── default.md
├── content/               # 站点内容(Markdown)
│   ├── _index.md          # 首页正文
│   ├── about.md           # 独立页面:关于
│   └── posts/              # 文章区段
│       ├── _index.md
│       ├── welcome-to-hugo.md
│       ├── markdown-guide.md
│       └── static-site-benefits.md
├── data/                  # 自定义数据文件(yaml/json/toml),当前为空
├── layouts/                # 模板层
│   ├── _default/
│   │   ├── baseof.html     # 基础骨架模板(head/header/footer + block)
│   │   ├── list.html       # 列表页模板(文章列表/标签/分类页复用)
│   │   └── single.html     # 详情页模板(文章正文/独立页面复用)
│   ├── partials/
│   │   ├── head.html       # <head> 部分,含 title/meta/css 引入
│   │   ├── header.html     # 顶部导航
│   │   └── footer.html     # 页脚
│   └── index.html          # 首页专属模板
├── static/                 # 原样拷贝到输出目录的静态资源
│   ├── css/style.css
│   └── images/
├── hugo.toml                # 站点主配置文件
├── .gitignore
└── README.md
```

## 环境要求

需要本地安装 Hugo（推荐 **extended** 版本，因为用到了 Sass/Scss 相关能力预留）。

- macOS: `brew install hugo`
- Windows: `choco install hugo-extended` 或 `winget install Hugo.Hugo.Extended`
- Linux: 从 [Hugo Releases](https://github.com/gohugoio/hugo/releases) 下载对应平台的 `hugo_extended_*` 压缩包，解压后将 `hugo` 二进制放入 PATH

验证安装：

```bash
hugo version
```

## 启动运行

在项目根目录下执行：

```bash
# 本地开发预览(带热重载),默认 http://localhost:1313
hugo server -D

# 若需要局域网内其他设备访问
hugo server -D --bind 0.0.0.0

# 生产环境构建,静态文件输出到 public/ 目录
hugo --minify
```

构建完成后，`public/` 目录下就是可以直接部署到任意静态托管服务（GitHub Pages / Cloudflare Pages / Netlify / Nginx 等）的完整 HTML/CSS 文件集合。

## 新建一篇文章

```bash
hugo new posts/my-new-post.md
```

会基于 `archetypes/default.md` 在 `content/posts/` 下生成一篇带 Front Matter 的新文章，`draft: true` 默认不会出现在正式构建中，本地预览时用 `hugo server -D` 才能看到草稿。

## 页面类型说明

| 页面 | 对应内容文件 | 使用的模板 |
| --- | --- | --- |
| 首页 | `content/_index.md` | `layouts/index.html` |
| 文章列表页 `/posts/` | `content/posts/_index.md` | `layouts/_default/list.html` |
| 文章详情页 | `content/posts/*.md` | `layouts/_default/single.html` |
| 独立页面(关于) | `content/about.md` | `layouts/_default/single.html` |

## 已本地验证

该工程已使用 `hugo v0.140.2 extended` 完整构建验证通过（`hugo --minify` 成功生成 `public/` 目录，无报错），可直接下载后按上方步骤启动。
