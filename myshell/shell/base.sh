#!/bin/bash
#This script displays the date and who's logged on

#如果想在同一行显示
#echo -n -e 'The time is:\n\n'
function zero() {
    echo The time is:
    date
    echo The one who has been logged is:
    who
}


function bc() {
    echo "----------------"
    #!/bin/bash
    var1=100
    var2=25
    var3=`echo "scale=4; $var1 / $var2" | bc`
    echo The answer for this is $var3
}
function one() {
#  one x y
    echo "第一个参数：$1";echo "第二个参数：$2";
}
function three() {
    echo "User info fro userId:$USER"
    echo UID:$UID
    echo HOME:$HOME
    #换行
    echo -e '\n'
    echo 'The cost of the item is \$15'
}

function four() {
    today=`date +%y%m%d`
    ls /usr/bin -al > log.$today
}

function five() {
    #!/bin/bash
    #An example of using the expr command

    var1=10
    var2=20
    var3=`expr $var2 + $var1`
    echo "The result is $var3"
}

function main() {
    at  -f test.sh now
}
main

