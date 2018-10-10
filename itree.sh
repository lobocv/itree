#!/bin/bash

source ${HOME}/.config/itree/preferences

DIR=$(itree2 "$@")

if [ -d "$DIR" ]; then
	echo "Changing directory to $DIR"
	cd $DIR
	if [ "${PrintDirOnExit}" = "1" ]; then
		ls
	fi	
fi


