cmd_dir=`dirname $0`
cmd_name=`basename $0`
if [ "`echo $cmd_dir | cut -c1`" != "/" ] ; then
    cmd_dir="`pwd`/$cmd_dir"
fi
base_dir="`echo $cmd_dir | sed -e s=/dev_tools==`/gocode"
cd $base_dir

log() {
  TS=`date +%Y/%m/%d" "%H:%M:%S",000"`
  echo "$TS | $1 | $cmd_name: $2"
}

error() {
  log ERROR "$1"
}

warn() {
  log WARNING "$1"
}

info() {
  log INFO "$1"
}

run_command() {
    info "run command \"$*\""
    "$@" || (error "while running command \"$*\"" && return 1)
}

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
  mkdir -p $base_dir/build && \
  cover_package internal && \
  cover_package ufpath && \
  cover_package joule && \
  cover_package cabritbx && \
  cover_package plumber && \
  cover_package testfs --tags=test_testfs && \
  cover_package cabridss --tags=test_testfs,test_cabridss && \
  cover_package cabrisync --tags=test_testfs,test_cabridss && \
  cover_package cabriui && \
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
