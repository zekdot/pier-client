# It's where you work, you should also put this script here.
export WORKSPACE=$HOME/bitxhub1.6.5
# It's fabric-samples's path. fabric-samples is cloned from GitHub.
export FABRIC_SAMPLE_PATH=$WORKSPACE/fabric-samples
# It's our project's path
export PIER_CLIENT_PATH=$WORKSPACE/pier-client

# start the fabric-network
function run_network() {
    cd $FABRIC_SAMPLE_PATH/test-network
    ./network.sh up createChannel
}
function prepare_env_var() {
    # export CHAINCODE_DIR=/home/zekdot/experiment
    export FABRIC_CFG_PATH=$FABRIC_SAMPLE_PATH/config/
    export CORE_PEER_TLS_ENABLED=true
    export ORDERER_CA=$FABRIC_SAMPLE_PATH/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    export PEER0_ORG1_CA=$FABRIC_SAMPLE_PATH/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
    export PEER0_ORG2_CA=$FABRIC_SAMPLE_PATH/test-network/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
    export PEER0_ORG3_CA=$FABRIC_SAMPLE_PATH/test-network/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
    export ORDERER_ADMIN_TLS_SIGN_CERT=$FABRIC_SAMPLE_PATH/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt
    export ORDERER_ADMIN_TLS_PRIVATE_KEY=$FABRIC_SAMPLE_PATH/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.key

    export CHANNEL_NAME=mychannel

    export CC_NAME=$1
    export CC_SRC_PATH=.
    export CC_SRC_LANGUAGE=go
    export CC_RUNTIME_LANGUAGE=golang
    export CC_VERSION=1.0
    export CC_SEQUENCE=1
    export INIT_REQUIRED=--init-required
}
function switch_orgs1() {
    export CORE_PEER_LOCALMSPID="Org1MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG1_CA
    export CORE_PEER_MSPCONFIGPATH=$FABRIC_SAMPLE_PATH/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
    export CORE_PEER_ADDRESS=localhost:7051
}
function switch_orgs2() {
    export CORE_PEER_LOCALMSPID="Org2MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG2_CA
    export CORE_PEER_MSPCONFIGPATH=$FABRIC_SAMPLE_PATH/test-network/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
    export CORE_PEER_ADDRESS=localhost:9051
}
# install related chaincode
function install_chaincode() {
    cd $FABRIC_SAMPLE_PATH/test-network
    
    cp $PIER_CLIENT_PATH/example/contracts/src/${CC_NAME}/${CC_NAME}.tar.gz .
    switch_orgs1
    peer lifecycle chaincode install ${CC_NAME}.tar.gz
    switch_orgs2
    peer lifecycle chaincode install ${CC_NAME}.tar.gz
}
function get_package_id() {
    peer lifecycle chaincode queryinstalled | sed 's|Package ID:\ ||g' | grep ${CC_NAME} | awk 'BEGIN {FS=", "} {print $1}'
}
function approve_chaincode() {
    switch_orgs1
    echo orgs1 approving the chaincode
    peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "$ORDERER_CA" --channelID $CHANNEL_NAME --name ${CC_NAME} --version ${CC_VERSION} --package-id ${PACKAGE_ID} --sequence ${CC_SEQUENCE} ${INIT_REQUIRED} ${CC_END_POLeICY} ${CC_COLL_CONFIG}
    switch_orgs2
    echo orgs2 approving the chaincode
    peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "$ORDERER_CA" --channelID $CHANNEL_NAME --name ${CC_NAME} --version ${CC_VERSION} --package-id ${PACKAGE_ID} --sequence ${CC_SEQUENCE} ${INIT_REQUIRED} ${CC_END_POLeICY} ${CC_COLL_CONFIG}
}
function commit_chaincode() {
    switch_orgs1
    peer lifecycle chaincode commit -o localhost:7050 \
    --ordererTLSHostnameOverride orderer.example.com --tls \
    --cafile "$ORDERER_CA" --channelID $CHANNEL_NAME --name ${CC_NAME} \
    --peerAddresses localhost:7051 --tlsRootCertFiles \
    $FABRIC_SAMPLE_PATH/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses localhost:9051 --tlsRootCertFiles \
    $FABRIC_SAMPLE_PATH/test-network/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    --version ${CC_VERSION} --sequence ${CC_SEQUENCE} ${INIT_REQUIRED} ${CC_END_POLICY} ${CC_COLL_CONFIG}
    peer lifecycle chaincode querycommitted --channelID $CHANNEL_NAME --name ${CC_NAME}
}
function init_chaincode() {
    peer chaincode invoke -o localhost:7050 \
    --ordererTLSHostnameOverride orderer.example.com \
    --tls --cafile "$ORDERER_CA" -C $CHANNEL_NAME \
    -n broker --peerAddresses localhost:7051 \
    --tlsRootCertFiles $FABRIC_SAMPLE_PATH/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses localhost:9051 \
    --tlsRootCertFiles $FABRIC_SAMPLE_PATH/test-network/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    --isInit -c '{"function":"Init","Args":[]}'
}
function prepare_chaincode() {
    prepare_env_var $1
    install_chaincode
    export PACKAGE_ID=`get_package_id`
    approve_chaincode
    commit_chaincode
    init_chaincode
}
function start() {
    run_network
    prepare_chaincode broker
    # prepare_chaincode data_swapper
}
start