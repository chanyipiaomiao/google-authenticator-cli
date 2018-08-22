## google-authenticator-cli

Google Authenticator 命令行客户端

## 安装使用

直接[**下载**](https://github.com/chanyipiaomiao/google-authenticator-cli/releases)二级制文件即可

## 使用说明

#### 添加

手动添加
```sh
google-authenticator-cli add --name="aliyun" --secret=xxxxxxxxxxx

--name   指定的是标识，就是为了方便辨认
--secret 一般跟两步验证二维码一起展示，方便手动添加的
```

识别二维码添加

```sh
./google-authenticator-cli add --name="Test" --qrcode=二维码图片路径
```

#### 删除

```sh
google-authenticator-cli delete --delete-secret=xxxxxxxxxxx

--delete-secret 和添加功能里面一样
```

#### 展示6位数字

展示所有

```sh
google-authenticator-cli show
```

```sh
Common 603175 18
第一列 标识
第二列 6位数字
第三列 剩余的时间(秒) 30秒循环一次
```

展示指定的secret

```sh
google-authenticator-cli show --show-secret="xxxx"
```

#### 可以使用watch命令动态显示

```sh
watch -n 1 google-authenticator-cli show
```
![示例](demo/google-autherticator.gif)