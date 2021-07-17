#! /bin/bash
# shellcheck disable=SC2046
# stop all DFS component including client
kill -9 `ps -ef | grep 'NodeRunner' | awk '{print $2}'`
