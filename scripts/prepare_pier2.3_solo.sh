# It's where you work, you should also put this script here.
export WORKSPACE=$HOME/bitxhub1.6.5
# BitXHub's official pier-client project cloned from GitHub.
export PIER_CLIENT_PATH=${WORKSPACE}/pier-client

export BITXHUB_ADDRESS=localhost

function compile_plugins() {
    cd $PIER_CLIENT_PATH/plugin
    # git checkout v1.6.5
    make general
}


function prepare_pier_toml() {
    head -n 27 $HOME/.pier/pier.toml > $HOME/.pier/pier.toml.new
    echo "addrs = [\"$BITXHUB_ADDRESS:60011\"]" >> $HOME/.pier/pier.toml.new
    echo 'timeout_limit = "1s"' >> $HOME/.pier/pier.toml.new
    echo "quorum = 2" >> $HOME/.pier/pier.toml.new
    echo "validators = [" >> $HOME/.pier/pier.toml.new
    cat $WORKSPACE/bitxhub.toml  | grep -E -o 0x[^\"]* | xargs -I {} echo \"{}\", >> $HOME/.pier/pier.toml.new
    tail -n -19 $HOME/.pier/pier.toml | head -n -3 >> $HOME/.pier/pier.toml.new
    echo "[appchain]" >> $HOME/.pier/pier.toml.new
    echo 'plugin = "general.so"' >> $HOME/.pier/pier.toml.new
    echo 'config = "general"' >> $HOME/.pier/pier.toml.new
    mv $HOME/.pier/pier.toml.new $HOME/.pier/pier.toml
}

function prepare_pier() {
    # 清除历史文件
    rm -rf $HOME/.pier
    # # 生成配置
    pier --repo=$HOME/.pier init
    # 根据bitxhub.toml生成对应的pier.toml
    prepare_pier_toml
    # 创建插件文件夹并进行拷贝
    mkdir $HOME/.pier/plugins
    cp $PIER_CLIENT_PATH/plugin/build/general $HOME/.pier/plugins/general.so
    cp -r $PIER_CLIENT_PATH/scripts/config $HOME/.pier/fabric
    # # 准备加密材料
    cp -r $PIER_CLIENT_PATH/scripts/crypto-config $HOME/.pier/fabric/
    # 复制Fabric上验证人证书
    cp $HOME/.pier/fabric/crypto-config/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/msp/signcerts/peer1.org2.example.com-cert.pem $HOME/.pier/fabric/fabric.validators

    # 修改网络配置和路径
    sed -i "s:\${CONFIG_PATH}:${HOME}/.pier/fabric:g" $HOME/.pier/fabric/config.yaml
    sed -i 's/host.docker.internal/localhost/g' $HOME/.pier/fabric/config.yaml
    # 可能需要修改$HOME/.pier/fabric/fabric.toml，这里只替换一下第一行的内容就行
    cat $PIER_CLIENT_PATH/scripts/config/fabric.toml | sed '1c addr = "localhost:7053"' > $HOME/.pier/fabric/fabric.toml
    # 最后对中继链进行注册
    pier --repo $HOME/.pier appchain register \
        --name=hf232 \
        --type=fabric \
        --consensusType RAFT \
        --validators=$HOME/.pier/fabric/fabric.validators \
        --desc="chainB-description" \
        --version=1.0.0
# bitxhub --repo ~/bitxhub1.6.5/bitxhub/scripts/build_solo client governance vote --id 0x41042c31e64575567C99a887351B406523466260-0 --info approve --reason approve
        # pier --repo ~/.pier rule deploy --path=/home/xy/.pier/fabric/validating.wasm
    # 部署验证规则
    # pier rule deploy --path $FABRIC_RULE_PATH
    export PIER_ID=`pier --repo=$HOME/.pier id`
    echo PIER_ID is $PIER_ID
    # echo "pier is ready, address is $PIER_ID, you need to save this value for send cross-chain request"
    # echo "run following code to start pier"
    # echo "pier --repo=$HOME/.pier start"
}
function start() {
    # 编译插件
    compile_plugins
    # 准备插件
    prepare_pier
}
start
# prepare_pier_toml