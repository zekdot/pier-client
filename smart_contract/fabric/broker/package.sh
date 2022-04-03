PEER_FULL_PATH='/media/zekdot/Linux_Workspace/fabric学习/fabric-samples/bin/peer'
function package_chaincode() {
    go mod tidy
    go mod vendor
    export CC_NAME=$1
    export CC_SRC_PATH=.
    export CC_SRC_LANGUAGE=go
    export CC_RUNTIME_LANGUAGE=golang
    export CC_VERSION=1.0
    $PEER_FULL_PATH lifecycle chaincode package ${CC_NAME}.tar.gz --path ${CC_SRC_PATH} --lang ${CC_RUNTIME_LANGUAGE} --label ${CC_NAME}_${CC_VERSION}
}
package_chaincode broker