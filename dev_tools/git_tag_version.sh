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
if [ "$GTV_AUTO" ] ; then
  run_command git add -A && \
  run_command git commit  && \
  run_command git push && \
  true || (info failed && exit 1)
fi
true && \
  [ -z "`git status -s`" ] || (error "`echo ; git status -s`" && exit 1) && \
  run_command git tag -f $current_version && \
  run_command git push origin $current_version -f && \
  true || (info failed && exit 1)
st=$?
info "ended"
exit $st
