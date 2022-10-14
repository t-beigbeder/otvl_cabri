cmd_dir=`dirname $0`
cmd_name=`basename $0`
if [ "`echo $cmd_dir | cut -c1`" != "/" ] ; then
    cmd_dir="`pwd`/$cmd_dir"
fi
base_dir="`echo $cmd_dir | sed -e s=/dev_tools==`/webui"
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
PATH="$HOME/go/bin:$PATH"
st=0
info "starting"
true && \
  yarn build && \
  true || (info failed && exit 1)
st=$?
info "ended"
exit $st
