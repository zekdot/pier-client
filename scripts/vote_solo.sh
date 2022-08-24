# It's where you work, you should also put this script here.
export WORKSPACE=$HOME/bitxhub1.6.5
for i in {1..3};
do
    $HOME/go/bin/bitxhub --repo $WORKSPACE/bitxhub/scripts/certs/node$i \
        client governance vote \
        --id $1 \
        --info approve --reason approve
done
