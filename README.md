# spider项目介绍

下载缓存及本地查看(喵绅士)漫画win工具，使用golang + lxn/walk 开发

## 使用

* 下载Release( https://github.com/baibaidexue/spider/releases )解压
* 双击spider.exe运行
* 程序自动监听剪切板变化，询问是否下载（可在setting页面取消勾选clipborad listen） ||  手动粘贴地址至主窗体-new下载


## 运行环境
仅发布针对于windows的版本

X86平台win10调测运行正常，其余win版本的兼容性未知，可以通过实际运行来验证

## 漫画存放目录

* 当启动spider.exe后，有漫画开始下载时，会在spider程序同样的路径创建star目录用于存放漫画资源
* star目录下会根据漫画名称划分不同的存放文件夹，该漫画的所有的资源文件均会在各自的目录下存放

## 程序主界面

### 程序启动后默认界面

![main](https://user-images.githubusercontent.com/84616906/131244648-984a7371-491d-4146-b1bb-187973cd3ae6.png)

### 已下载漫画预览

![image](https://user-images.githubusercontent.com/84616906/131247493-140d36a3-9fde-4aa2-92db-bff03ac4a67d.png)


### 漫画管理界面使用右键点击漫画图标执行操作


_____________________
## 弹出式下载窗口页功能


![downloadinfo](https://user-images.githubusercontent.com/84616906/123549412-794e2800-d79b-11eb-97e4-28f3955f786e.png)

#### 功能按钮使用说明

* metainfo repull
    - 重新下载漫画元数据（重新下载封面，重写下载信息至md下载详情记录文件）
* images repull
    - 重新下载manga中缺失的图片（因为网络原因导致的下载失败会重新下载，或者不完整的图片，手动删除后再次点击会重新下载缺失的图片）
* zip images
    - 打包当前的漫画图片至zip压缩包，用于漫画阅读器（下载完成后默认自动打包，可在主界面配置）
* 下载进度及下载过程输出信息
* open manga's folder
    - 打开当前窗口正在下载的漫画的基目录
