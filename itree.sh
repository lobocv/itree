#!/bin/bash

source ${HOME}/.config/itree/preferences

# Execute itree and capture the resulting command it spits back
CMD=$(itree2 "$@")

# Execute the command it spits back
eval ${CMD}

