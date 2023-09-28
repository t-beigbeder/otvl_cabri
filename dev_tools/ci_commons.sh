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

backup_error() {
  btd=/tmp/betc-${test_count}-$PID
  mkdir $btd
  cp -a $BTD/. $btd/
  info "find current status here: $btd"
  return 0
}

run_silent() {
  if [ "$LCMD" ]; then
    echo $* > $LCMD
  fi
  "$@" > $OUT 2> $ERR || (error "while running command \"$*\"" && cat $OUT $ERR && backup_error && return 1)
}

run_error() {
    "$@" && (error "error expected while running command \"$*\"" && return 1)
    return 0
}

run_bg_cmd() {
  info "run background command \"$*\""
  "$@" &
  pidc=$!
  if [ $? -ne 0 ] ; then
    error "while running command \"$*\""
    return 1
  fi
  info "pidc $pidc"
  return 0
}

run_bg_silent() {
  "$@" &
  pidc=$!
  if [ $? -ne 0 ] ; then
    error "while running command \"$*\""
    return 1
  fi
  return 0
}

find_out() {
  grep "$@" $OUT > /dev/null || (error "command `cat $LCMD` didn't produce $*" && cat $OUT && return 1 )
}

get_out() {
  SHOUT=`sha256sum < $OUT`
  LCOUT=`wc -l < $OUT`
  WCOUT=`wc -c < $OUT`
}
