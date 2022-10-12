# 简介

pier-client是基于BitXHub实现的可以支持多种区块链的通用化插件。

# 目录结构
```
├── broker_service 不同应用链的客户端实现
├── cli_client 默认的命令行客户端实现
├── plugin 插件逻辑的实现
├── scripts 部署相关shell脚本
├── send_many_interchainGet 发送并发交易的工具
└── smart_contract 智能合约
```

# 使用
Before start. We assume your take a directory as a workspace. In my situation, workspace is $HOME/bitxhub1.6.5。Put fabric-samples and pier-client and pier under workspace. Then copy all shell scripts in pier-client/scripts to your workspace. The workspace should be look like this.

```
├── fabric-samples
├── pier
├── pier-client
├── prepare_pier2.3_solo.sh
└── run_fabric2.3.sh
```

## How to build in Fabric2.3

### Start Fabric2.3

Use run_fabric2.3.sh to start fabric2.3 network.

Before execute it. You need modify three enviroment variable at the begining of file.

* WORKSPACE your workspace
* FABRIC_SAMPLE_PATH your fabric-samples dir, if you workspace look like mine, then you don't have to change it.
* PIER_CLIENT_PATH pier-client location, if you workspace look like mine, then you don't have to change it.

Run following command to start fabric and deploy smart contract we need.

```sh
bash run_fabric2.3.sh
```

When seeing this, fabric2.3 is ready.

```
384ef440fc341189995eedc92a1ebcd3859613bd1052d7dc44e057ac] committed with status (VALID) at localhost:9051
2022-04-08 13:51:11.266 UTC [chaincodeCmd] ClientWait -> INFO 002 txid [1c1a35a4384ef440fc341189995eedc92a1ebcd3859613bd1052d7dc44e057ac] committed with status (VALID) at localhost:7051
Committed chaincode definition for chaincode 'broker' on channel 'mychannel':
Version: 1.0, Sequence: 1, Endorsement Plugin: escc, Validation Plugin: vscc, Approvals: [Org1MSP: true, Org2MSP: true]
2022-04-08 13:51:14.086 UTC [chaincodeCmd] chaincodeInvokeOrQuery -> INFO 001 Chaincode invoke successful. result: status:200 
```

### Start broker_service

Open a new terminal to enter pier-client/broker_service/fabric, run following command to start broker_service.

```sh
rm -rf keystore wallet # execute every time you restart your fabric2.3
go env -w GO111MODULE=on
go run *.go
```

When seeing following output, broker_service is ready.

```
2022/04/08 13:57:06 ============ application-golang starts ============
 [fabsdk/core] 2022/04/08 13:57:06 UTC - cryptosuite.GetDefault -> INFO No default cryptosuite found, using default SW implementation
2022/04/08 13:57:06 start listen
```

### Compile command-line

Enter pier-client/cli_client and run following command.

```sh
go build
```

When seeing a executable program called cli. Run following command to init bitxhub's meta-data in fabric2.3.

```sh
./cli init
```

At the same time, you can see your broker_service terminal will output following contents:

```
2022/04/08 13:57:06 start listen
set inner-meta to {}
set outter-meta to {}
set callback-meta to {}
```

### Start pier

Use prepare_pier2.3_solo.sh to start a pier. Make sure your bitxhub is in solo mode. Before execute it, modify following variable at the beginning of this file.

* WORKSPACE modify to your workspace, should be same as in run_fabric2.3.sh
* PIER_CLIENT_PATH your fabric-samples dir, if you workspace look like mine, then you don't have to change it.
* BITXHUB_ADDRESS ip address that you deploy your bitxhub relay chain.

Run following command to register your pier to bitxhub and get proposal id.

```sh
bash prepare_pier2.3_solo.sh
```

You should get following method:

```sh
NFO[14:12:39.012] Establish connection with bitxhub 172.19.241.113:60011 successfully  module=rpcx
the register request was submitted successfully, chain id is 0xA814C52DeB8FEa9F0De9fbaC01b8Cf70f581F611, proposal id is 0xA814C52DeB8FEa9F0De9fbaC01b8Cf70f581F611-0
PIER_ID is 0xA814C52DeB8FEa9F0De9fbaC01b8Cf70f581F611
```

As show in output, proposal id is 0xA814C52DeB8FEa9F0De9fbaC01b8Cf70f581F611-0. Use it to vote and then deploy validation_rule.

Then start the pier.

```sh
pier start
```

When seeing this, pier is ready.

```
3 index=0 module=syncer
INFO[14:15:15.863] Handle interchain tx wrapper                  count=0 height=4 index=0 module=syncer
INFO[14:15:15.863] Handle interchain tx wrapper                  count=0 height=5 index=0 module=syncer
INFO[14:15:15.863] Handle interchain tx wrapper                  count=0 height=6 index=0 module=syncer
INFO[14:15:15.863] Handle interchain tx wrapper                  count=0 height=7 index=0 module=syncer
INFO[14:15:15.863] Syncer started                                bitxhub_height=7 current_height=7 module=syncer
INFO[14:15:15.888] Exchanger started 
```

Now whole network is ready. You can use it now.

## How to use in Fabric2.3

All operation will use cli in pier-client/cli_client.

### Set value

Get value1 to key1 in fabric2.3

```sh
./cli set key1 value1
```

### Get value

Get value of key1

```sh
./cli get key1
```

output will be like this.

```
key1 :  value1
```

### InterchainGet with Ethereum

Assuming pierId is 0xA814C52De856Ea9F0De9fbaC01b8Cf70f581F611, your contract address is 0xEA452DeB8FEa9F0De9fbaC01b8Cf754581F6118. If you want to get value of key2 in Ethereum, use following method

==If you want to fetch fabric data, contract address should be mychannel&data_swapper==

```sh
./cli interchainGet 0xA814C52De856Ea9F0De9fbaC01b8Cf70f581F611 0xEA452DeB8FEa9F0De9fbaC01b8Cf754581F6118
key2
```

Then see what pier will output to make sure transaction is valid. After maybe 10 seconds. Use this command to get actual value of key2.

```sh
./cli get key2
```

