#!/bin/bash

CWD=$(dirname $(readlink -e $0))
INSTALL_PATH=/usr/local/bin/itree.sh
BIN_INSTALL_PATH=/usr/local/bin/itree2
CUR_SHELL=$(basename $SHELL)
ITREE_ALIAS="alias itree=\". itree.sh\""
GOEXEC=$(which go)


echo "Detected shell $CUR_SHELL"
while ! [[ "$continue" =~ (Y|y|yes|N|n|no) ]]; do
    read -p "Do you want to install itree for $CUR_SHELL? [Y/n]" continue
done

if [[ "$continue" =~ (N|n|no) ]]; then
    read -p "Please type the name of the shell you would like to install itree with:" CUR_SHELL
    if [ -z $CUR_SHELL ]; then
        echo "Aborting installation."
        exit 0
    fi

fi

RC="${HOME}/.${CUR_SHELL}rc"

if [ ! -f ${RC} ]; then
    echo "Cannot find rc file ${RC}. Aborting installation."
    exit 1
fi

echo "Installing to ${INSTALL_PATH}"
sudo ${GOEXEC} build -o ${BIN_INSTALL_PATH} ${CWD}/itree.go
sudo cp ${CWD}/itree.sh ${INSTALL_PATH}

ALIAS_EXISTS=$(grep "${ITREE_ALIAS}" ${RC})
ERR=$?
if [ ${ERR} == 0 ]; then
    echo "itree wrapper already exists in ${RC}. Doing nothing."
else
    echo "Adding itree alias to ${RC}"
    echo "${ITREE_ALIAS}" >> ${RC}
fi


