moudle=$1
if [ -z ${moudle} ];then
    echo "need moudle [cli|srv] "
    exit 1
fi

op=$2

build() {
    go build -o dist/stp${moudle} ./cmd/stp${moudle}
}

pack() {
    cd dist
    version=$(./stp${moudle} -v)
    if [ ${moudle}x = "cli"x ];then
        tar -zcvf stpcli-${version}.tar.gz stpcli
        rm -f stpcli
    elif [ ${moudle}x = "srv"x ];then
        cp ../cmd/stpsrv/cfg.json . &&
        tar -zcvf stpsrv-${version}.tar.gz stpsrv cfg.json &&
        rm -f stpsrv cfg.json
    fi
    echo $version
    cd ..
}

if [ ${op}x = "p"x ];then
    build && pack
    echo "build and pack"
else
    build
    echo "just build"
fi
