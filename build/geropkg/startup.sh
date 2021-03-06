#!/bin/sh

ROOT=$(cd `dirname $0`; pwd)

sh ${ROOT}/stop.sh

export DYLD_LIBRARY_PATH=${ROOT}/czero/lib/
export LD_LIBRARY_PATH=${ROOT}/czero/lib/
DEFAULT_DATD_DIR="${ROOT}/data"
LOGDIR="${ROOT}/log"
DEFAULT_RPCPORT=8545
DEFAULT_PORT=60602

cmd="${ROOT}/bin/gero --datadir=${DEFAULT_DATD_DIR} --port ${DEFAULT_PORT}"
if [[ $# -gt 0 ]]; then
     while [[ "$1" != "" ]]; do
       	 case "$1" in
		--datadir)
		    cmd=${cmd/--datadir=${DEFAULT_DATD_DIR}/--datadir=$2};shift 2;;
        --dev)
		    cmd="$cmd --dev";shift;;

        --alpha)
		    cmd="$cmd --alpha";shift;;
        --rpc)
		    localhost=$(hostname -I|awk -F ' ' '{print $1}')
		    cmd="$cmd --rpc --rpcport $2 --rpcaddr $localhost --rpcapi personal,sero,web3 --rpccorsdomain '*'";shift 2;;
        --port)
            cmd=${cmd/--port ${DEFAULT_PORT}/--port $2};shift 2;;
        --keystore)
            cmd="$cmd --keystore $2";shift 2;;
		*)exit;;
        esac
    done
fi

mkdir -p ${ROOT}/log

echo $cmd
${cmd} &> ${ROOT}/log/gero.log & echo $! > ${ROOT}/pid