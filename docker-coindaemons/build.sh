BUILD="ari boli btg dash doge flo img ltc pac sum xsh zen arrr bsv dcr emc2 fto imgc mona pirl veles xvg axe btc dcrwallet emrals gto kmd moon pyrk vtc zch bca btcv dgb euno hana kzc nah rvn xlt zcl bch btcz cann dgc fch hatch lcc nsd sbtc xsg zec"

if [ ! -z "$1" ]; then
    BUILD="$1"
fi

echo "Building: [$BUILD]"

if [ -z "$PREFIX" ]; then
    PREFIX="pooldetective-coind"
fi

TAG_PREFIX=$PREFIX
if [ ! -z "$DOCKER_REGISTRY" ]; then
    TAG_PREFIX="$DOCKER_REGISTRY/$PREFIX"
fi


for prod in $BUILD 
do
    echo "Building $prod"
    PREPARESCRIPT="$PWD/$prod/prepare.sh"
    if [ -f "$PREPARESCRIPT" ]; then
        cd $prod
        source "$PREPARESCRIPT"
        cd ..
    fi

    docker build "./$prod/" -t "$TAG_PREFIX-$prod"
    if [ ! -z "$DOCKER_REGISTRY" ]; then
        docker push "$TAG_PREFIX-$prod"
    fi
done