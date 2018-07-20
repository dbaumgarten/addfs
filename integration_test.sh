#!/bin/bash

sourcedir=/tmp/addfs
mountp=/tmp/appendmount
origdir=`pwd`

cleanup(){
    cd /
    sudo killall -q addfs
    sudo umount $mountp
    sudo rm -rf $sourcedir
    sudo rm -rf $mountp
    sudo rm /tmp/lastcmd
}

fail(){
    echo Tests failed
    exit 1
}

expect_denied(){
`eval $@ &> /tmp/lastcmd`
if [ $? == 0 ]; then
    echo "Command \"$@\" should have failed but did not."
    cat /tmp/lastcmd
    fail
fi
}

expect_allowed(){
`eval $@ &> /tmp/lastcmd`
if [ $? -ne 0 ]; then
    echo "Command \"$@\" should have worked but did not."
    cat /tmp/lastcmd
    fail
fi
}

setup(){
echo "Testing ./addfs $sourcedir $mountp $@"
sudo mkdir -p $sourcedir 
mkdir -p $mountp
cd $origdir
sudo ./addfs $@ $sourcedir $mountp &
disown
sleep 1
sudo chmod 777 $mountp
cd $mountp
echo "test" > testfile
echo "test123" > testfileb
mkdir testdir
echo "test123" > testdir/testfileb
echo "test" > permtest
}

go build && chmod +x addfs

cleanup &> /dev/null
setup

expect_denied rm testfile
expect_denied rm testfileb
expect_denied rm -rf testdir
expect_denied rm testdir/testfileb
expect_denied 'echo "123" | tee testfile'
expect_denied 'echo "123" | tee -a testfile'
expect_denied 'echo "123" > testdir/testfileb'
expect_denied truncate testfile
expect_denied mv testfile other
expect_denied chown root:root testfile
expect_allowed chmod 777 testfile
expect_allowed cat testfile
expect_allowed cat testdir/testfileb

expect_denied sudo rm testfile
expect_denied sudo rm testfileb
expect_denied sudo rm -rf testdir
expect_denied sudo rm testdir/testfileb
expect_denied sudo 'echo "123" | sudo tee testfile'
expect_denied sudo 'echo "123" | sudo tee -a testfile'
expect_denied sudo truncate testfile
expect_denied sudo mv testfile other
expect_allowed sudo chown root:root testfile
expect_allowed sudo chmod 777 testfile
expect_allowed sudo cat testfile
expect_allowed sudo cat testdir/testfileb

expect_allowed chmod 000 permtest
expect_denied cat permtest

cleanup

setup --allowRootMutation

expect_denied rm testfile
expect_denied rm testfileb
expect_denied rm -rf testdir
expect_denied rm testdir/testfileb
expect_denied 'echo "123" > testfile'
expect_denied 'echo "123" >> testfile'
expect_denied 'echo "123" > testdir/testfileb'
expect_denied chown root:root testfile
expect_allowed ls
expect_allowed chmod 777 testfile
expect_allowed cat testfile
expect_allowed cat testdir/testfileb

expect_allowed sudo 'echo "123" | sudo tee testfile'
expect_allowed sudo 'echo "123" | sudo tee -a testfile'
expect_allowed sudo chown root:root testfile
expect_allowed sudo chmod 777 testfile
expect_allowed sudo cat testfile
expect_allowed sudo cat testdir/testfileb
expect_allowed sudo truncate testfile --size 0
expect_allowed sudo rm testfile
expect_allowed sudo rm testdir/testfileb
expect_allowed sudo rm -rf testdir
expect_allowed sudo mv testfileb other

cleanup

setup --mutableFiles '.*\.tmp$'

expect_denied rm testfile
expect_denied rm testfileb
expect_denied rm -rf testdir
expect_denied rm testdir/testfileb
expect_denied 'echo "123" | tee testfile'
expect_denied 'echo "123" | tee -a testfile'
expect_denied 'echo "123" > testdir/testfileb'
expect_denied truncate testfile
expect_denied mv testfile other
expect_denied chown root:root testfile
expect_allowed chmod 777 testfile
expect_allowed cat testfile
expect_allowed cat testdir/testfileb

expect_allowed 'echo "123aaaa" | tee foo.tmp'
expect_allowed 'echo "123aaaa" | tee foo.tmp'
expect_allowed 'echo "123aaaa" | tee -a foo.tmp'
expect_allowed truncate foo.tmp --size 2
expect_allowed mv foo.tmp bar.tmp
expect_allowed cp bar.tmp bar
expect_allowed rm bar.tmp
expect_denied rm bar

expect_denied sudo rm testfile
expect_denied sudo rm testfileb
expect_denied sudo rm -rf testdir
expect_denied sudo rm testdir/testfileb
expect_denied sudo 'echo "123" | sudo tee testfile'
expect_denied sudo 'echo "123" | sudo tee -a testfile'
expect_denied sudo truncate testfile
expect_denied sudo mv testfile other
expect_allowed sudo chown root:root testfile
expect_allowed sudo chmod 777 testfile
expect_allowed sudo cat testfile
expect_allowed sudo cat testdir/testfileb

cleanup

echo "Tests OK! Cleaning up..."






