cmd_dir=`dirname $0`
cmd_name=`basename $0`
if [ "`echo $cmd_dir | cut -c1`" != "/" ] ; then
    cmd_dir="`pwd`/$cmd_dir"
fi
root_dir="`echo $cmd_dir | sed -e s=/dev_tools==`"
current_version=`cat $root_dir/dev_tools/current_version.txt`

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