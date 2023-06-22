cmd_dir=`dirname $0`
cmd_name=`basename $0`
if [ "`echo $cmd_dir | cut -c1`" != "/" ] ; then
    cmd_dir="`pwd`/$cmd_dir"
fi
. $cmd_dir/ci_commons.sh
. $cmd_dir/build_commons.sh
base_dir="`echo $cmd_dir | sed -e s=/dev_tools==`/gocode"
cd $base_dir


cover_package() {
  cd $base_dir/packages/$1 && \
  # go vet . 2> $base_dir/build/$1.vet.out >&1 && \
  go test $2 -coverprofile=$base_dir/build/$1.coverage.out . && \
  go tool cover -html=$base_dir/build/$1.coverage.out -o $base_dir/build/$1.coverage.html
}

PATH="$HOME/go/bin:$PATH"
st=0
info "starting"
true && \
  write_go_version && \
  mkdir -p $base_dir/build && \
  cover_package internal && \
  cover_package ufpath && \
  cover_package joule && \
  cover_package cabrifsu && \
  cover_package cabritbx --tags=test_testfs,test_cabridss && \
  cover_package plumber && \
  cover_package testfs --tags=test_testfs && \
  cover_package cabridss --tags=test_testfs,test_cabridss && \
  cover_package cabrisync --tags=test_testfs,test_cabridss && \
  cover_package cabriui --tags=test_testfs,test_cabridss && \
  mkdir -p $base_dir/build $base_dir/cmds/locsv/frontend_build && \
  touch $base_dir/cmds/locsv/frontend_build/dummy && \
  cd $base_dir/cmds/locsv && \
  goimports -w . && \
  #go vet . 2> $base_dir/build/cmd.vet.out >&1 && \
  go build -o $base_dir/build/locsv ./main.go && \
  GOOS=windows GOARCH=amd64 go build -o $base_dir/build/locsv.exe ./main.go && \
  cd $base_dir/cabri && \
  goimports -w . && \
  #go vet . 2> $base_dir/build/cabri.vet.out >&1 && \
  go build -o $base_dir/build/cabri ./main.go && \
  GOOS=windows GOARCH=amd64 go build -o $base_dir/build/cabri.exe ./main.go && \
  true || (info failed && exit 1)
st=$?
info "ended"
exit $st
