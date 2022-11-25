# clone项目

创建工作目录bitxhub1.6.5，进入bitxhub1.6.5

```sh
cd $HOME/bitxhub1.6.5
```

clone三个项目。

pier

```sh
git clone https://github.com/meshplus/pier.git
cd pier
git checkout v1.6.5
make install
```

bitxhub

```sh
git clone https://github.com/meshplus/bitxhub.git
cd bitxhub
git checkout v1.6.5
```

pier-client

```sh
git clone 
cd pier-client
git checkout v1.0.0
```

# 保证HTTP地址可用

假设后台地址为`http://localhost:3005/getCustomPackage`

测试是否能从该地址获取到对应的结果

```sh
curl http://localhost:3005/getCustomPackage
```

查看是否能得到对应输出

```json
{"customPackage":{"agent":"SINTEF","catchPackageId":"17","consumerPackageId":"5","packingDate":"12321112"},"catchPackage":{"catchPackageId":"15","packingDate":"12321118","palletNum":"1"},"pallet":{"palletNum":"1","palletWeight":20.5,"productNum":"1","supplierId":"1","tripNo":"10"},"trip":{"departureDate":"2022/08/05","departurePort":"Trondheim","landingDate":"2022/08/15","landingPort":"Burgen","tripNo":"10","tripWithinYearNo":"240","vesselName":"Vessel"}}
```

# 启动broker_service

## 启动应用链客户端

进入`http`目录

```sh
cd $HOME/bitxhub1.6.5/pier-client/appchain_client/http
```

启动客户端

```sh
rm *_test.go
go run *.go
```

当出现如下内容时说明启动成功

```sh
~/bitxhub1.6.5/pier-client/appchain_client/http$ go run *.go
2022/11/24 23:49:38 start listen
```

## 启动broker_service

进入`broker_service`目录

```sh
cd $HOME/bitxhub1.6.5/pier-client/broker_service
```

修改`constants.go`，将`DB_PATH`修改为`$HOME/.pier/meta-data`

```go
package main

const (
	DB_PATH = "/home/hzh/.pier/meta-data"
	RPC_PORT = "1212"
	APPCHAIN_ADDRESS = "127.0.0.1:1211"
)
```

然后启动

```
rm *_test.go
go run *.go
```

出现如下内容说明启动成功

```sh
~/bitxhub1.6.5/broker_service$ go run *.go
2022/11/24 23:41:55 start listen
```

可以使用自带客户端验证是否有效

```sh
cd $HOME/bitxhub1.6.5/pier-client/cli_client
./cli get 5
```

如出现如下的输出，说明启动成功

```sh
~/bitxhub1.6.5/pier-client/cli_client$ ./cli get 5
5 :  {"customPackage":{"agent":"SINTEF","catchPackageId":"17","consumerPackageId":"5","packingDate":"12321112"},"catchPackage":{"catchPackageId":"15","packingDate":"12321118","palletNum":"1"},"pallet":{"palletNum":"1","palletWeight":20.5,"productNum":"1","supplierId":"1","tripNo":"10"},"trip":{"departureDate":"2022/08/05","departurePort":"Trondheim","landingDate":"2022/08/15","landingPort":"Burgen","tripNo":"10","tripWithinYearNo":"240","vesselName":"Vessel"}}
```

# 启动Bitxhub

进入`bitxhub`目录然后启动。

```sh
cd $HOME/bitxhub1.6.5/bitxhub
make solo
```

出现如下内容说明启动成功：

```
ule=executor
INFO[2022-11-25T00:50:10.368] BlockExecutor started                         hash=0x0BF41c03bE038e389470dBdC9ed53DF897F178921520AF37877188d1759A7068 height=1 module=executor
INFO[2022-11-25T00:50:10.368] Router module started                         module=router
INFO[2022-11-25T00:50:10.468] Order is ready                                module=app plugin_path=plugins/solo.so

=======================================================
    ____     _    __    _  __    __  __            __
   / __ )   (_)  / /_  | |/ /   / / / /  __  __   / /_
  / __  |  / /  / __/  |   /   / /_/ /  / / / /  / __ \
 / /_/ /  / /  / /_   /   |   / __  /  / /_/ /  / /_/ /
/_____/  /_/   \__/  /_/|_|  /_/ /_/   \__,_/  /_.___/

=======================================================
```

完成之后将目录下的`scripts/build_solo/bitxhub.toml`复制到工作目录下。

```sh
cp $HOME/bitxhub1.6.5/bitxhub/scripts/build_solo/bitxhub.toml $HOME/bitxhub1.6.5/bitxhub.toml
```

# 请求注册pier

进入脚本目录

```sh
cd $HOME/bitxhub1.6.5/pier-client/scripts/
```

其中`prepare_pier2.3_solo.sh`为网关启动脚本，`BITXHUB_ADDRESS`变量为bitxhub节点地址。

```sh
bash prepare_pier2.3_solo.sh
```

当输出如下信息，说明提交了注册请求

```
~/bitxhub1.6.5/pier-client/scripts$ bash prepare_pier2.3_solo.sh 
mkdir -p build
GO111MODULE=on go build -o build/general ./*.go
INFO[15:45:05.835] Establish connection with bitxhub localhost:60011 successfully  module=rpcx
the register request was submitted successfully, chain id is 0x68700a8505f4Cd07aa336e3A4C939E60Bff969CD, proposal id is 0x68700a8505f4Cd07aa336e3A4C939E60Bff969CD-0
PIER_ID is 0x68700a8505f4Cd07aa336e3A4C939E60Bff969CD
```

这里需要记住`proposal id`的值，如本例中是`0x68700a8505f4Cd07aa336e3A4C939E60Bff969CD-0`，下一步将会用到。

# BitXHub进行投票操作统同意注册

进入脚本目录

```sh
cd $HOME/bitxhub1.6.5/pier-client/scripts/
```

执行如下命令进行投票，参数需要修改为上一步的`proposal id`的值。

```sh
bash vote_solo.sh 0x68700a8505f4Cd07aa336e3A4C939E60Bff969CD-0
```

当出现如下输出说明注册请求已同意。

```sh
~/bitxhub1.6.5/pier-client/scripts$ bash vote_solo.sh 0x7E6D93E642802fb276Ba0735d7289E2E1a45455B-0
vote successfully!
vote successfully!
vote successfully!
```

# 启动pier

启动pier

```sh
pier start
```

当输出如下内容时，说明启动成功

```
INFO[15:52:16.772] Persist block header                          height=15 module=bxh_lite
INFO[15:52:16.772] BitXHub lite started                          bitxhub_height=15 current_height=15 module=bxh_lite
INFO[15:52:16.772] Syncer recover                                begin=15 end=15 module=syncer
INFO[15:52:16.773] X:Before handleInterchainWrapperAndPersist    module=syncer
INFO[15:52:16.773] Z:!!!!!!!!!!Start to execute handleInterchainTxWrapper  count=0 height=0 index=0 module=syncer
INFO[15:52:16.773] Z:!!!!!!!!!!Before verifyWrapper it           count=0 height=0 index=0 module=syncer
INFO[15:52:16.773] Z:!!!!!!!!!!After verifyWrapper it            count=0 height=0 index=0 module=syncer
INFO[15:52:16.773] Z:!!!!!!!!!!End to execute handleInterchainTxWrapper  count=0 height=0 index=0 module=syncer
INFO[15:52:16.773] Handle interchain tx wrapper                  count=0 height=15 index=0 module=syncer
INFO[15:52:16.773] X:After handleInterchainWrapperAndPersist     module=syncer
INFO[15:52:16.773] Syncer started                                bitxhub_height=15 current_height=15 module=syncer
INFO[15:52:16.775] Exchanger started                             module=exchanger
```