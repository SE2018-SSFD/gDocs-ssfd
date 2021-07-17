#! /bin/bash
# shellcheck disable=SC2046
kill -9 `ps -ef | grep 'NodeRunner client' | awk '{print $2}'`
