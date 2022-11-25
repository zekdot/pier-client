# It's where you work, you should also put this script here.
export WORKSPACE=$HOME/bitxhub1.6.5
if [ $# -lt 1 ];
then
  echo usage: $0 pierId
  exit;
fi
for i in {1..3};
do
    $GOPATH/bin/bitxhub --repo $WORKSPACE/bitxhub/scripts/certs/node$i \
        client governance vote \
        --id $1 \
        --info approve --reason approve
done
