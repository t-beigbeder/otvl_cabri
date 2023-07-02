cmd_dir=`dirname $0`
cmd_name=`basename $0`
if [ "`echo $cmd_dir | cut -c1`" != "/" ] ; then
    cmd_dir="`pwd`/$cmd_dir"
fi
. $cmd_dir/ci_commons.sh
. $cmd_dir/build_commons.sh
base_dir="`echo $cmd_dir | sed -e s=/dev_tools==`/gocode"
cd $base_dir

PATH="$HOME/go/bin:$PATH"
st=0
info "starting"
true && \
  write_go_version && \
  mkdir -p $base_dir/build && \
  mkdir -p $base_dir/build $base_dir/cmds/locsv/frontend_build && \
  touch $base_dir/cmds/locsv/frontend_build/dummy && \
  cd $base_dir/cmds/locsv && \
  goimports -w . && \
  go vet . 2> $base_dir/build/cmds.vet.out >&1 && \
  CGO_ENABLED=0 go build -o $base_dir/build/locsv ./main.go && \
  GOOS=windows GOARCH=amd64 go build -o $base_dir/build/locsv.exe ./main.go && \
  cd $base_dir/cabri && \
  goimports -w . && \
  go vet . 2> $base_dir/build/cabri.vet.out >&1 && \
  CGO_ENABLED=0 go build -o $base_dir/build/cabri ./main.go && \
  GOOS=windows GOARCH=amd64 go build -o $base_dir/build/cabri.exe ./main.go && \
  distribute_bin $base_dir/build/cabri $base_dir/build/cabri.exe && \
  true || (info failed && exit 1)
st=$?
info "ended"
exit $st
