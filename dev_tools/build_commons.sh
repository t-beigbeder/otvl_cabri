current_version=v0.1.0
cmd_dir=`dirname $0`
cmd_name=`basename $0`
if [ "`echo $cmd_dir | cut -c1`" != "/" ] ; then
    cmd_dir="`pwd`/$cmd_dir"
fi
root_dir="`echo $cmd_dir | sed -e s=/dev_tools==`"

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

write_go_version() {
  vf=$root_dir/gocode/packages/cabridss/version.go
  echo "package cabridss" > $vf
  echo const CabriVersion = "\"$current_version `git show --no-patch --no-notes --pretty='%h %cd'`\"" >> $vf
}

distribute_bin() {
  for p in `echo $CABRI_DIST_PATH | sed -e 's=:= =g' ` ; do
    for f in $* ; do
      cp -p $f $p/
    done
  done
}