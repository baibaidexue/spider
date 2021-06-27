# spider项目介绍
pull comics from zhahentai(喵绅士)

使用golang + lxn/walk 开发的喵绅士简易漫画下载工具

## 运行环境
仅发布针对于windows的版本

X86平台win10调测运行正常，其余win版本的兼容性未知，可以通过实际运行来验证

## 工作目录说明

* 当启动spider.exe后，有漫画开始下载时，会在spider程序同样的路径创建star目录用于存放漫画资源
* star目录下会根据漫画名称划分不同的存放文件夹，该漫画的所有的资源文件均会在各自的目录下存放

## 漫画下载的说明

### 程序启动后主界面

程序启动后出现如下窗口，（也许还会有一个cmddebug输出窗口）

![spidermain](https://user-images.githubusercontent.com/84616906/123549407-75220a80-d79b-11eb-8775-1f94a194f6f0.png)

### 下载方式
* 手动模式：手动在主界面粘贴喵绅士漫画链接下载
* 自动模式：程序启动当剪切板内容变更，如果被鉴别可能是喵绅士漫画链接，会提示是否下载该链接

点击下载后，漫画的存放目录为当前spider程序所在的star目录（**不建议更改目录**，在当前目录如果没有star目录程序会自动创建）


### 下载窗口页介绍


![downloadinfo](https://user-images.githubusercontent.com/84616906/123549412-794e2800-d79b-11eb-97e4-28f3955f786e.png)

#### 功能按钮使用说明

* metainfo repull
    - 重新下载漫画元数据（重新下载封面，重写下载信息至md下载详情记录文件）
* images repull
    - 重新下载manga中缺失的图片（因为网络原因导致的下载失败会重新下载，或者不完整的图片，手动删除后再次点击会重新下载缺失的图片）
* zip images
    - 打包当前的漫画图片至zip压缩包，用于漫画阅读器（下载完成后自动打包）
* 下载进度及下载过程输出信息
* open manga's folder
    - 打开当前窗口正在下载的漫画的基目录
