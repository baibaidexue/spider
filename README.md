# spider项目介绍

基于爱好驱动与golang语言实践，闲时开发的下载缓存及本地查看(喵绅士)漫画win工具，使用golang + lxn/walk库完成代码实现


## 下载与使用

* 下载地址：Release页面( https://github.com/baibaidexue/spider/releases )，zip包解压后运行spider.exe文件
* 如何开启下载：
     * 默认开启剪切板监听，判断后询问是否下载（该功能可在setting页面关闭 “clipboard listen ”）
     * 手动将URL地址粘贴主窗体->new->url 点击开始下载
* 集成简单的本地漫画管理功能，管理star目录的已下载漫画文件（由于每次都会从本地目录加载，启动会耗费2s~5s左右）
* 每页数据加载量可通过漫画管理界面左上角下拉选项调整，reload后应用至页面，默认值10


## 运行环境
仅有windows平台版本（lxn/walk库限制）

x86_x64平台win10/win11调测运行正常，其余win版本的兼容性未知

## 漫画存放

* 自动创建star目录用于存放漫画资源
* star目录下根据**漫画名**划分不同的存放位置

## 程序主界面

### 程序启动后默认启动至本地漫画管理界面

![main](https://user-images.githubusercontent.com/84616906/131244648-984a7371-491d-4146-b1bb-187973cd3ae6.png)

### 右键漫画缩略图开启管理菜单

![image](https://user-images.githubusercontent.com/84616906/131247493-140d36a3-9fde-4aa2-92db-bff03ac4a67d.png)

#### Tips
* 鼠标停留在漫画缩略图时会显示漫画本地信息，包含图片数量，磁盘空间占用大小

_____________________
## 弹出式下载窗口页功能


![downloadinfo](https://user-images.githubusercontent.com/84616906/123549412-794e2800-d79b-11eb-97e4-28f3955f786e.png)

#### 功能按钮使用说明

* metainfo repull
    - 重新下载漫画元数据（重新下载封面，重写下载信息至md下载详情记录文件）
* images repull
    - 重新下载manga中缺失的图片（因为网络原因导致的下载失败会重新下载，或者不完整的图片，手动删除后再次点击会重新下载缺失的图片）
* zip images
    - 打包当前的漫画图片至zip压缩包，用于漫画阅读器（默认自动打包，可在setting中更改配置）
* 下载进度及下载过程输出信息
* open manga's folder
    - 使用Explorer打开该漫画的本地目录
