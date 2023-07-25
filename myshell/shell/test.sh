#!/bin/bash
set -o errexit  #遇到错误就停止  set -e
set -o pipefail  #当多个|分割的管道命令其中有一个出错就当做失败，如果不设置就会按最后一个的结果来判定成功或者失败
set -o nounset   #不加nounset参数时，遇到不存在的变量，默认情况下会忽略，直接执行下行语句,加上之后就会提示错误  set -u
set -o xtrace  #这将打开命令跟踪。虽然命令本身不输出任何内容，但后续命令将在执行前打印出来。    set -x

ROOT=$(cd $(dirname ${BASH_SOURCE[0]}) && pwd -P)
IMAGE_FILE=${IMAGE_FILE:-"tkestack.io/gaia/vcuda:latest"}   #参考readme
echo $ROOT
echo $IMAGE_FILE




