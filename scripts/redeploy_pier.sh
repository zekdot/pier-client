pier-client/cli_client/cli init
proposalId=`bash prepare_pier2.3_solo.sh | egrep -o "0x[A-Za-z0-9]+-0"`
echo proposalId is $proposalId
ssh xy@xy bash /home/xy/bitxhub1.6.5/vote_solo.sh $proposalId
pier --repo $HOME/.pier rule deploy --path=$HOME/.pier/fabric/validating.wasm
pier start 2>&1 | tee pier.out