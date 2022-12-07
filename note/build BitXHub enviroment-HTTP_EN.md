# Clone project and install it

Create workspace dir `bitxhub1.6.5` and enter it

```sh
mkdir $HOME/bitxhub1.6.5
cd $HOME/bitxhub1.6.5
```

There is three projects to be cloned

pier

```sh
cd $HOME/bitxhub1.6.5
git clone https://github.com/meshplus/pier.git
cd pier
git checkout v1.6.5
make install
```

bitxhub

```sh
cd $HOME/bitxhub1.6.5
git clone https://github.com/meshplus/bitxhub.git
cd bitxhub
git checkout v1.6.5
```

pier-client

```sh
cd $HOME/bitxhub1.6.5
git clone 
cd pier-client
git checkout v1.0.0
```

# Make sure url is available

Assume that url of backend is `http://localhost:3005/getCustomPackage`

Use `curl` command to make sure url is available

```sh
curl http://localhost:3005/getCustomPackage
```

If we can get following output, then we can go to next step.

```json
{"customPackage":{"agent":"SINTEF","catchPackageId":"17","consumerPackageId":"5","packingDate":"12321112"},"catchPackage":{"catchPackageId":"15","packingDate":"12321118","palletNum":"1"},"pallet":{"palletNum":"1","palletWeight":20.5,"productNum":"1","supplierId":"1","tripNo":"10"},"trip":{"departureDate":"2022/08/05","departurePort":"Trondheim","landingDate":"2022/08/15","landingPort":"Burgen","tripNo":"10","tripWithinYearNo":"240","vesselName":"Vessel"}}
```

# Start broker_service

## Start appchain-api client

Enter `http` dir.

```sh
cd $HOME/bitxhub1.6.5/pier-client/appchain_client/http
```

Start client

```sh
rm *_test.go
go run *.go
```

When we get following output, start is successful.

```
~/bitxhub1.6.5/pier-client/appchain_client/http$ go run *.go
2022/11/24 23:49:38 start listen
```

## Start broker_service

Enter `broker_service` dir.

```sh
cd $HOME/bitxhub1.6.5/pier-client/broker_service
```

Modify `constants.go`, change `DP_PATH` to your version, assume your username in linux is `hzh`, then content will be.

```go
package main

const (
	DB_PATH = "/home/hzh/.pier/meta-data"
	RPC_PORT = "1212"
	APPCHAIN_ADDRESS = "127.0.0.1:1211"
)
```

Then we can start it.

```sh
rm *_test.go
go run *.go
```

If we have the following output, then broker_service start successfully.

```
~/bitxhub1.6.5/broker_service$ go run *.go
2022/11/24 23:41:55 start listen
```

We can confirm that by using `cli` command. 

```sh
cd $HOME/bitxhub1.6.5/pier-client/cli_client
./cli get 5
```

If we have following output, then the start is successful.

```
~/bitxhub1.6.5/pier-client/cli_client$ ./cli get 5
5 :  {"customPackage":{"agent":"SINTEF","catchPackageId":"17","consumerPackageId":"5","packingDate":"12321112"},"catchPackage":{"catchPackageId":"15","packingDate":"12321118","palletNum":"1"},"pallet":{"palletNum":"1","palletWeight":20.5,"productNum":"1","supplierId":"1","tripNo":"10"},"trip":{"departureDate":"2022/08/05","departurePort":"Trondheim","landingDate":"2022/08/15","landingPort":"Burgen","tripNo":"10","tripWithinYearNo":"240","vesselName":"Vessel"}}
```

# Start Bitxhub

Enter `bitxhub` dir and start it.

```sh
cd $HOME/bitxhub1.6.5/bitxhub
make solo
```

If we have following output, then the start is successful.

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

After that, copy `bitxhub.toml` to your workspace.

```sh
cp $HOME/bitxhub1.6.5/bitxhub/scripts/build_solo/bitxhub.toml $HOME/bitxhub1.6.5/bitxhub.toml
```

# Request to register pier info to bitxhub node

Enter `scripts` dir 

```sh
cd $HOME/bitxhub1.6.5/pier-client/scripts/
```

If you bitxhub node and your pier not deploy in the same instance, you need to modify `prepare_pier2.3_solo.sh` and change `BITXHUB_ADDRESS` according your situation.

```sh
bash prepare_pier2.3_solo.sh
```

If we have following output, then the request is successful.

```
~/bitxhub1.6.5/pier-client/scripts$ bash prepare_pier2.3_solo.sh 
mkdir -p build
GO111MODULE=on go build -o build/general ./*.go
INFO[15:45:05.835] Establish connection with bitxhub localhost:60011 successfully  module=rpcx
the register request was submitted successfully, chain id is 0x68700a8505f4Cd07aa336e3A4C939E60Bff969CD, proposal id is 0x68700a8505f4Cd07aa336e3A4C939E60Bff969CD-0
PIER_ID is 0x68700a8505f4Cd07aa336e3A4C939E60Bff969CD
```

Here we need to write down `proposal id` in the example, it's `0x68700a8505f4Cd07aa336e3A4C939E60Bff969CD-0`, we will use it in the following steps.

# Agree request

Enter `scripts` dir

```sh
cd $HOME/bitxhub1.6.5/pier-client/scripts/
```

Execute the following command to execute vote process and agree the request. In the following command, `0x68700a8505f4Cd07aa336e3A4C939E60Bff969CD-0` is the `proposal id` we get in previous step.

```sh
bash vote_solo.sh 0x68700a8505f4Cd07aa336e3A4C939E60Bff969CD-0
```

If we have following output, then the agree is successful.

```
~/bitxhub1.6.5/pier-client/scripts$ bash vote_solo.sh 0x7E6D93E642802fb276Ba0735d7289E2E1a45455B-0
vote successfully!
vote successfully!
vote successfully!
```

# Start pier

Deploy validation file.

```sh
pier --repo $HOME/.pier rule deploy --path=$HOME/.pier/fabric/validating.wasm
```

Start pier

```sh
pier start
```

Initiate the meta-data

```sh
cd $HOME/bitxhub1.6.5/pier-client/cli_client
./cli init
```

If we have following output, then the start is successful.

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



Then the pier in HTTP method is ready now.