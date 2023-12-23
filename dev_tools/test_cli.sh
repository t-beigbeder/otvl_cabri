cmd_dir=`dirname $0`
cmd_name=`basename $0`
if [ "`echo $cmd_dir | cut -c1`" != "/" ] ; then
    cmd_dir="`pwd`/$cmd_dir"
fi
. $cmd_dir/ci_commons.sh
. $cmd_dir/build_commons.sh
root_dir="`echo $cmd_dir | sed -e s=/dev_tools==`"
PATH="$root_dir/gocode/build:$PATH"
test_count=0
# on windows set WBT_PATH to windows temp path like
# /Users/ziggy/AppData/Local/Temp
tmp_dir=/tmp
if [ "${WBT_PATH}" ]; then
  tmp_dir=${WBT_PATH}
fi

setup_test() {
  PID=$$
  test_count=`expr $test_count + 1`
  TD=${tmp_dir}/tc-${test_count}-$PID
  BTD=/tmp/tc-${test_count}-$PID
  export PID TD BTD
  trap "cd /tmp; chmod -R +w /tmp/tc-*-$PID; rm -rf /tmp/tc-*-$PID" 1 2 3 15 EXIT
  mkdir -p $BTD && \
    mkdir $BTD/tmp && \
    cd $BTD && \
    HOME=$TD && \
    LCMD=$BTD/tmp/lcmd && \
    ALLCMD=$BTD/tmp/allcmd && \
    OUT=$BTD/tmp/out && \
    ERR=$BTD/tmp/err && \
    cabri cli config --get && \
    cabri cli config --gen u1 u2 && \
    mkdir ${BTD}/fsyb1 ${BTD}/fsyb2 && \
    true
}

untar_simple() {
  tar xzf $root_dir/tests/data/simple.tar.gz
}

untar_advanced() {
  tar xzf $root_dir/tests/data/advanced.tar.gz
}

update_advanced() {
  adv=$1 && \
  chmod +w $adv/d4/d41ro && date >  $adv/d4/d41ro/f411 && mkdir $adv/d4/d41ro/d412 && \
  chmod -w  $adv/d4/d41ro && \
  chmod +w $adv/d4/d42/d421ro && date >>  $adv/d4/d42/d421ro/f4211rw && \
  chmod +w $adv/d4/d42/d421ro/f4212ro && date >> $adv/d4/d42/d421ro/f4212ro && chmod -w $adv/d4/d42/d421ro/f4212ro && \
  chmod -w $adv/d4/d42/d421ro && \
  true
}

update_acl() {
  dir=$1 && \
  date >> $dir/d1/d11/f3 && \
  date >> $dir/d1/d11/f3bis && \
  true
}

make_olf() {
  dir=$1
  olf=$2
  wc=$3
  if [ "$wc" ] ; then
    mkdir $BTD/wc || return 1
    cdir="--cdir $wc"
  else
    cdir=
  fi
  mkdir $dir && \
  run_silent cabri cli $cdir dss make $olf -s s
}

make_polf() {
  dir=$1
  olf=$2
  wc=$3
  if [ "$wc" ] ; then
    mkdir $BTD/wc || return 1
    cdir="--cdir $wc"
  else
    cdir=
  fi
  mkdir $dir && \
  run_silent cabri cli $cdir dss make $olf -s s --ximpl bdb
}

make_obs() {
  dir=$1
  obs=$2
  wc=$3
  if [ "$wc" ] ; then
    mkdir $BTD/wc || return 1
    cdir="--cdir $wc"
  else
    cdir=
  fi
  mkdir $dir && \
  run_silent cabri cli $cdir dss make $OBS_ENV $obs && \
#  run_command cabri cli check $OBS_ENV --s3ls && \
  run_silent cabri cli $cdir dss clean $obs && \
  true
}

run_test_gp() {
  run_silent cabri cli dss get $dest@d1/f1 $TD/simple/d1/d11/f1clone && \
  run_silent cmp $TD/simple/d1/f1 $TD/simple/d1/d11/f1clone && \
  run_silent cabri cli dss updns $dest@d1/d11 -c f1clone --acl x-uid:1000: --acl : && \
  run_silent cabri cli dss put $dest@d1/d11/f1clone $TD/simple/d1/d11/f1clone --acl x-uid:1000: --acl : && \
  run_silent cabri cli dss get $dest@d1/d11/f1clone $TD/simple/d1/d11/f1clone2 && \
  run_silent cmp $TD/simple/d1/d11/f1clone $TD/simple/d1/d11/f1clone2 && \
  true
}

run_basic_sync() {
  ori=$1
  dest=$2
  test_gp=$3
  run_silent cabri cli sync $ori@ $dest@ -rd && \
  find_out "created: 15" && \
  run_silent cabri cli sync $ori@ $dest@ -rv && \
  find_out "created: 15" && \
  run_silent cabri cli sync $ori@ $dest@ -rd && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  ([ -z "$test_gp" ] || run_test_gp) && \
  true
}

run_advanced_sync() {
  ori=$1
  dest=$2
  adv=$3
  run_silent cabri cli sync $ori@ $dest@ -rd && \
  find_out "created: 23" && \
  run_silent cabri cli sync $ori@ $dest@ -rv && \
  find_out "created: 23" && \
  run_silent cabri cli sync $ori@ $dest@ -rd && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  ([ "$test_cli_fast" ] || sleep 2) && \
  update_advanced $adv && \
  run_silent cabri cli sync $ori@ $dest@ -rd && \
  find_out "created: 2, updated 3," && \
  run_silent cabri cli sync $ori@ $dest@ -rv --summary && \
  find_out "created: 2, updated 3," && \
  true
}

run_acl_sync() {
  ori=$1
  dest=$2
  simple=$3
  fsyb1=fsy:${TD}/fsyb1
  fsyb2=fsy:${TD}/fsyb2
  run_silent cabri cli sync $ori@ $dest@ --acl u1: --acl u2:rx --macl :u1 --macl :u2 -u u1 -rd && \
  find_out "created: 13, updated 1, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $ori@ $dest@ --acl u1: --acl u2:rx --macl :u1 --macl :u2 -u u1 -rv && \
  find_out "created: 13, updated 1, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $ori@ $dest@ --acl u1: --acl u2:rx --macl :u1 --macl :u2 -u u1 -rv && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $dest@ $fsyb1@ --macl u1: -rd && \
  find_out "created: 13, updated 1, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $dest@ $fsyb1@ --macl u1: -rv && \
  find_out "created: 13, updated 1, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $dest@ $fsyb1@ --macl u1: -rd && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $dest@ $fsyb2@ --macl u2: -rd && \
  find_out "created: 13, updated 1, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $dest@ $fsyb2@ --macl u2: -rv && \
  find_out "created: 13, updated 1, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $dest@ $fsyb2@ --macl u2: -rd && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  update_acl $simple && \
  run_silent cabri cli sync $ori@ $dest@ --acl u1: --acl u2:rx --macl :u1 --macl :u2 -u u1 -rd && \
  find_out "created: 1, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $ori@ $dest@ --acl u1: --acl u2:rx --macl :u1 --macl :u2 -u u1 -rv && \
  find_out "created: 1, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $ori@ $dest@ --acl u1: --acl u2:rx --macl :u1 --macl :u2 -u u1 -rv && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $dest@ $fsyb1@ --macl u1: -rd && \
  find_out "created: 1, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $dest@ $fsyb1@ --macl u1: -rv && \
  find_out "created: 1, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $dest@ $fsyb1@ --macl u1: -rd && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $dest@ $fsyb2@ --macl u2: -rd && \
  find_out "created: 1, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $dest@ $fsyb2@ --macl u2: -rv && \
  find_out "created: 1, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $dest@ $fsyb2@ --macl u2: -rd && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  true
}

run_basic_unlock() {
  dss=$1
  run_silent cabri cli dss mkns $dss@ --acl : > /dev/null && \
  run_silent cabri cli lsns $dss@ && \
  run_error cabri cli dss unlock --lock $dss 2> /dev/null && \
  run_error cabri cli lsns $dss@ 2> /dev/null && \
  run_silent cabri cli dss unlock $dss 2> /dev/null && \
  run_silent cabri cli lsns $dss@ && \
  true
}

run_index_err() {
  dss=$1
  run_error cabri cli dss audit $dss 2> /dev/null && \
  true
}

run_index() {
  dss=$1
  dssd=$2
  rdss=$3
  run_silent cabri cli dss audit $dss && \
  run_silent cabri cli dss scan $dss && \
  run_silent cabri cli dss lshisto -rs $dss@ && \
  get_out && sh1=$SHOUT && \
  run_silent cp $dssd/index.bdb $dssd/index.bdb.bck && \
  run_silent cp -a $HOME/.cabri $HOME/.cabri.bck && \
  ( [ "$rdss" ] || run_silent cabri cli dss reindex $dss) && \
  run_silent cabri cli dss lshisto -rs $dss@ && \
  get_out && [ "$SHOUT" = "$sh1" ] && \
#  backup_error && \
  true
}

test_basic_sync_olf() {
  info test_basic_sync_olf && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  olf=olf:${TD}/olf && \
  make_olf $BTD/olf $olf && \
  run_basic_sync $fsy $olf test_gp && \
  true
}

test_basic_sync_polf() {
  info test_basic_sync_polf && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  olf=olf:${TD}/olf && \
  make_polf $BTD/olf $olf && \
  run_basic_sync $fsy $olf test_gp && \
  true
}

test_basic_sync_xolf() {
  info test_basic_sync_xolf && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  olf=xolf:${TD}/olf && \
  make_olf $BTD/olf $olf && \
  run_basic_sync $fsy $olf test_gp && \
  true
}

test_basic_sync_obs() {
  info test_basic_sync_obs && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  obs=obs:${TD}/obs && \
  make_obs $BTD/obs $obs && \
  run_basic_sync $fsy $obs test_gp && \
#  run_command cabri cli check $OBS_ENV --s3ls && \
  true
}

test_basic_sync_xobs() {
  info test_basic_sync_xobs && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  obs=xobs:${TD}/obs && \
  make_obs $BTD/obs $obs && \
  run_basic_sync $fsy $obs test_gp && \
#  run_command cabri cli check $OBS_ENV --s3ls && \
  true
}

test_basic_sync_wolf() {
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  olf=olf:${TD}/olf && \
  make_polf $BTD/olf $olf $TD/wc && \
  run_bg_silent cabri webapi --cdir $TD/wc olf+http://localhost:3000/$TD/olf@wo && \
  sleep 1 && \
  wo=webapi+http://localhost:3000/wo && \
  run_basic_sync $fsy $wo test_gp && \
  run_silent kill $pidc && \
  true
}

test_basic_sync_wolf_noe() {
  info test_basic_sync_wolf_noe && \
  export CABRIDSS_WEB_RAISE_ERROR= && \
  test_basic_sync_wolf && \
  true
}

test_basic_sync_wolf_err() {
  info test_basic_sync_wolf_err && \
  export CABRIDSS_WEB_RAISE_ERROR=1 && \
  test_basic_sync_wolf && \
  true
}

test_basic_sync_wobs() {
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  obs=obs:${TD}/obs && \
  make_obs $BTD/obs $obs $TD/wc && \
  run_bg_silent cabri webapi --cdir $TD/wc obs+http://localhost:3000/$TD/obs@wo && \
  sleep 1 && \
  wo=webapi+http://localhost:3000/wo && \
  run_basic_sync $fsy $wo test_gp && \
  run_silent kill $pidc && \
  true
}

test_basic_sync_wobs_noe() {
  info test_basic_sync_wobs_noe && \
  export CABRIDSS_WEB_RAISE_ERROR= && \
  test_basic_sync_wobs && \
  true
}

test_basic_sync_wobs_err() {
  info test_basic_sync_wobs_err && \
  export CABRIDSS_WEB_RAISE_ERROR=1 && \
  test_basic_sync_wobs && \
  true
}

test_basic_sync_xwolf() {
  info test_basic_sync_xwolf && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  olf=xolf:${TD}/olf && \
  make_olf $BTD/olf $olf $TD/wc && \
  run_bg_silent cabri webapi --cdir $TD/wc xolf+http://localhost:3000/$TD/olf@wo && \
  sleep 1 && \
  wo=xwebapi+http://localhost:3000/wo && \
  run_basic_sync $fsy $wo test_gp && \
  run_silent kill $pidc && \
  true
}

test_basic_sync_xwobs() {
  info test_basic_sync_xwobs && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  obs=xobs:${TD}/obs && \
  make_obs $BTD/obs $obs $TD/wc && \
  run_bg_silent cabri webapi --cdir $TD/wc xobs+http://localhost:3000/$TD/obs@wo && \
  sleep 1 && \
  wo=xwebapi+http://localhost:3000/wo && \
  run_basic_sync $fsy $wo test_gp && \
  run_silent kill $pidc && \
#  run_command cabri cli check $OBS_ENV --s3ls && \
  true
}

test_basic_sync_wfs() {
  info test_basic_sync_wfs && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  run_silent mkdir -p ${TD}/simple2 && \
  wfs=xobs:${TD}/wfs && \
  mkdir $BTD/wfs $BTD/wc && \
  run_bg_silent cabri webapi --cdir $TD/wc fsy+http://localhost:3000/$TD/wfs@wfs && \
  sleep 1 && \
  wfs=wfsapi+http://localhost:3000/wfs && \
  run_basic_sync $fsy $wfs test_gp && \
  run_silent cabri cli sync $wfs@ fsy:${TD}/simple2@ -rdn && \
  find_out "created: 13, updated 1, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $wfs@ fsy:${TD}/simple2@ -rvn && \
  find_out "created: 13, updated 1, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $wfs@ fsy:${TD}/simple2@ -rdn && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent kill $pidc && \
  true
}

test_advanced_sync_olf() {
  info test_advanced_sync_olf && \
  setup_test && \
  untar_advanced && \
  adv=${TD}/advanced && \
  fsy=fsy:${TD}/advanced && \
  olf=olf:${TD}/olf && \
  make_olf $BTD/olf $olf && \
  run_advanced_sync $fsy $olf $adv && \
  true
}

test_advanced_sync_xolf() {
  info test_advanced_sync_xolf && \
  setup_test && \
  untar_advanced && \
  adv=${TD}/advanced && \
  fsy=fsy:${TD}/advanced && \
  olf=xolf:${TD}/olf && \
  make_olf $BTD/olf $olf && \
  run_advanced_sync $fsy $olf $adv && \
  true
}

test_sync_back_and_forth_olf() {
  info test_sync_back_and_forth_olf && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  run_silent cp -a ${TD}/simple ${TD}/simple2 && \
  run_silent date > ${TD}/simple2/d1/f2bis && \
  run_silent date >> ${TD}/simple2/d1/f2 && \
  run_silent date > ${TD}/simple2/d2/d21/f5bis && \
  olf=olf:${TD}/olf && \
  make_polf $BTD/olf $olf && \
  run_basic_sync $fsy $olf && \
  run_silent cabri cli sync fsy:${TD}/simple2@ $olf@ -rdn && \
  find_out "created: 2, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync fsy:${TD}/simple2@ $olf@ -rvn && \
  find_out "created: 2, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $olf@ $fsy@ -rdn && \
  find_out "created: 2, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $olf@ $fsy@ -rvn && \
  find_out "created: 2, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $olf@ $fsy@ -rdn && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync fsy:${TD}/simple2@ $fsy@ -rdn && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  true
}

test_sync_back_and_forth_xolf() {
  info test_sync_back_and_forth_xolf && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  run_silent cp -a ${TD}/simple ${TD}/simple2 && \
  run_silent date > ${TD}/simple2/d1/f2bis && \
  run_silent date >> ${TD}/simple2/d1/f2 && \
  run_silent date > ${TD}/simple2/d2/d21/f5bis && \
  olf=xolf:${TD}/olf && \
  make_olf $BTD/olf $olf && \
  run_basic_sync $fsy $olf && \
  run_silent cabri cli sync fsy:${TD}/simple2@ $olf@ -rdn && \
  find_out "created: 2, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync fsy:${TD}/simple2@ $olf@ -rvn && \
  find_out "created: 2, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $olf@ $fsy@ -rdn && \
  find_out "created: 2, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $olf@ $fsy@ -rvn && \
  find_out "created: 2, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $olf@ $fsy@ -rdn && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync fsy:${TD}/simple2@ $fsy@ -rdn && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  true
}

test_sync_back_and_forth_xwolf() {
  info test_sync_back_and_forth_xwolf && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  run_silent cp -a ${TD}/simple ${TD}/simple2 && \
  run_silent date > ${TD}/simple2/d1/f2bis && \
  run_silent date >> ${TD}/simple2/d1/f2 && \
  run_silent date > ${TD}/simple2/d2/d21/f5bis && \
  olf=xolf:${TD}/olf && \
  make_olf $BTD/olf $olf $TD/wc && \
  run_bg_silent cabri webapi --cdir $TD/wc xolf+http://localhost:3000/$TD/olf@wo && \
  sleep 1 && \
  wo=xwebapi+http://localhost:3000/wo && \
  run_basic_sync $fsy $wo && \
  run_silent cabri cli sync fsy:${TD}/simple2@ $wo@ -rdn && \
  find_out "created: 2, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync fsy:${TD}/simple2@ $wo@ -rvn && \
  find_out "created: 2, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $wo@ $fsy@ -rdn && \
  find_out "created: 2, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $wo@ $fsy@ -rvn && \
  find_out "created: 2, updated 2, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync $wo@ $fsy@ -rdn && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent cabri cli sync fsy:${TD}/simple2@ $fsy@ -rdn && \
  find_out "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" && \
  run_silent kill $pidc && \
  true
}

test_acl_sync_olf() {
  info test_acl_sync_olf && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  olf=olf:${TD}/olf && \
  simple=${TD}/simple && \
  make_olf $BTD/olf $olf && \
  run_acl_sync $fsy $olf $simple && \
  true

}

test_basic_unlock_olf() {
    info test_basic_unlock_olf && \
    setup_test && \
    olf=olf:${TD}/olf && \
    make_polf $BTD/olf $olf && \
    run_basic_unlock $olf
    true
}

test_basic_unlock_xolf() {
    info test_basic_unlock_xolf && \
    setup_test && \
    olf=xolf:${TD}/olf && \
    make_olf $BTD/olf $olf && \
    run_basic_unlock $olf
    true
}

test_basic_unlock_obs() {
  info test_basic_unlock_obs && \
  setup_test && \
  obs=obs:${TD}/obs && \
  make_obs $BTD/obs $obs && \
  run_basic_unlock $obs && \
  true
}

test_basic_unlock_xobs() {
  info test_basic_unlock_xobs && \
  setup_test && \
  obs=xobs:${TD}/obs && \
  make_obs $BTD/obs $obs && \
  run_basic_unlock $obs && \
  true
}

test_basic_unlock_wolf() {
    info test_basic_unlock_wolf && \
    setup_test && \
    olf=olf:${TD}/olf && \
    make_polf $BTD/olf $olf $TD/wc && \
    run_bg_silent cabri webapi --cdir $TD/wc olf+http://localhost:3000/$TD/olf@wo && \
    sleep 1 && \
    wo=webapi+http://localhost:3000/wo && \
    run_basic_unlock $wo && \
    run_silent kill $pidc && \
    true
}

test_basic_unlock_wobs() {
    info test_basic_unlock_wobs && \
    setup_test && \
    obs=obs:${TD}/obs && \
    make_obs $BTD/obs $obs $TD/wc && \
    run_bg_silent cabri webapi --cdir $TD/wc obs+http://localhost:3000/$TD/obs@wo && \
    sleep 1 && \
    wo=webapi+http://localhost:3000/wo && \
    run_basic_unlock $wo && \
    run_silent kill $pidc && \
    true
}

test_basic_unlock_xwolf() {
    info test_basic_unlock_xwolf && \
    setup_test && \
    olf=xolf:${TD}/olf && \
    make_olf $BTD/olf $olf $TD/wc && \
    run_bg_silent cabri webapi --cdir $TD/wc xolf+http://localhost:3000/$TD/olf@wo && \
    sleep 1 && \
    wo=xwebapi+http://localhost:3000/wo && \
    run_basic_unlock $wo && \
    run_silent kill $pidc && \
    true
}

test_basic_unlock_xwobs() {
    info test_basic_unlock_xwobs && \
    setup_test && \
    obs=xobs:${TD}/obs && \
    make_obs $BTD/obs $obs $TD/wc && \
    run_bg_silent cabri webapi --cdir $TD/wc xobs+http://localhost:3000/$TD/obs@wo && \
    sleep 1 && \
    wo=xwebapi+http://localhost:3000/wo && \
    run_basic_unlock $wo && \
    run_silent kill $pidc && \
    true
}

test_index_olf() {
  info test_index_olf && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  olf=olf:${TD}/olf && \
  make_olf $BTD/olf $olf && \
  run_basic_sync $fsy $olf && \
  run_index_err $olf && \
  true
}

test_index_polf() {
  info test_index_polf && \
  setup_test && \
  untar_advanced && \
  adv=${TD}/advanced && \
  fsy=fsy:${TD}/advanced && \
  olf=olf:${TD}/olf && \
  make_polf $BTD/olf $olf && \
  run_advanced_sync $fsy $olf $adv && \
  run_index $olf $TD/olf && \
  true
}

test_index_xolf() {
  info test_index_xolf && \
  setup_test && \
  untar_advanced && \
  adv=${TD}/advanced && \
  fsy=fsy:${TD}/advanced && \
  olf=xolf:${TD}/olf && \
  make_olf $TD/olf $olf && \
  run_advanced_sync $fsy $olf $adv && \
  run_index $olf $TD/olf && \
  true
}

test_index_obs() {
  info test_index_obs && \
  setup_test && \
  untar_advanced && \
  adv=${TD}/advanced && \
  fsy=fsy:${TD}/advanced && \
  obs=obs:${TD}/obs && \
  make_obs $BTD/obs $obs && \
  run_advanced_sync $fsy $obs $adv && \
  run_index $obs $TD/obs && \
  true
}

test_index_xobs() {
  info test_index_xobs && \
  setup_test && \
  untar_advanced && \
  adv=${TD}/advanced && \
  fsy=fsy:${TD}/advanced && \
  obs=xobs:${TD}/obs && \
  make_obs $BTD/obs $obs && \
  run_advanced_sync $fsy $obs $adv && \
  run_index $obs $TD/obs && \
  true
}

test_index_wolf() {
  info test_index_wolf && \
  setup_test && \
  untar_advanced && \
  adv=${TD}/advanced && \
  fsy=fsy:${TD}/advanced && \
  olf=olf:${TD}/olf && \
  make_polf $BTD/olf $olf $TD/wc && \
  run_bg_silent cabri webapi --cdir $TD/wc olf+http://localhost:3000/$TD/olf@wo && \
  sleep 1 && \
  wo=webapi+http://localhost:3000/wo && \
  run_advanced_sync $fsy $wo $adv && \
  run_index $wo $TD/olf 1 && \
  run_silent kill $pidc && \
  true
}

test_index_wobs() {
  info test_index_wobs && \
  setup_test && \
  untar_advanced && \
  adv=${TD}/advanced && \
  fsy=fsy:${TD}/advanced && \
  obs=obs:${TD}/obs && \
  make_obs $BTD/obs $obs $TD/wc && \
  run_bg_silent cabri webapi --cdir $TD/wc obs+http://localhost:3000/$TD/obs@wo && \
  sleep 1 && \
  wo=webapi+http://localhost:3000/wo && \
  run_advanced_sync $fsy $wo $adv && \
  run_index $wo $TD/obs 1 && \
  run_silent kill $pidc && \
#  backup_error && \
#  false && \
  true
}

test_index_xwolf() {
  info test_index_xwolf && \
  setup_test && \
  untar_advanced && \
  adv=${TD}/advanced && \
  fsy=fsy:${TD}/advanced && \
  olf=xolf:${TD}/olf && \
  make_olf $BTD/olf $olf $TD/wc && \
  run_bg_silent cabri webapi --cdir $TD/wc xolf+http://localhost:3000/$TD/olf@wo && \
  sleep 1 && \
  wo=xwebapi+http://localhost:3000/wo && \
  run_advanced_sync $fsy $wo $adv && \
  run_index $wo $TD/olf 1 && \
  run_silent kill $pidc && \
  true
}

test_index_xwobs() {
  info test_index_xwobs && \
  setup_test && \
  untar_advanced && \
  adv=${TD}/advanced && \
  fsy=fsy:${TD}/advanced && \
  obs=xobs:${TD}/obs && \
  make_obs $BTD/obs $obs $TD/wc && \
  run_bg_silent cabri webapi --cdir $TD/wc xobs+http://localhost:3000/$TD/obs@wo && \
  sleep 1 && \
  wo=xwebapi+http://localhost:3000/wo && \
  run_advanced_sync $fsy $wo $adv && \
  run_index $wo $TD/obs 1 && \
  run_silent kill $pidc && \
  true
}

test_serve_olf_xolf() {
  info test_serve_olf_xolf && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  olf=olf:${TD}/olf && \
  make_polf $BTD/olf $olf $TD/wc && \
  run_error cabri webapi --cdir $TD/wc xolf+http://localhost:3000/$TD/olf@wo 2> /dev/null && \
  run_bg_silent cabri webapi --cdir $TD/wc olf+http://localhost:3000/$TD/olf@wo && \
  sleep 1 && \
  run_silent kill $pidc && \
  true
}

test_serve_obs_xobs() {
  info test_serve_obs_xobs && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  obs=obs:${TD}/obs && \
  make_obs $BTD/obs $obs $TD/wc && \
  run_error cabri webapi --cdir $TD/wc xobs+http://localhost:3000/$TD/obs@wo 2> /dev/null && \
  run_bg_silent cabri webapi --cdir $TD/wc obs+http://localhost:3000/$TD/obs@wo && \
  sleep 1 && \
  run_silent kill $pidc && \
  true
}

test_serve_xolf_olf() {
  info test_serve_xolf_olf && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  olf=xolf:${TD}/olf && \
  make_polf $BTD/olf $olf $TD/wc && \
  run_error cabri webapi --cdir $TD/wc olf+http://localhost:3000/$TD/olf@wo 2> /dev/null && \
  run_bg_silent cabri webapi --cdir $TD/wc xolf+http://localhost:3000/$TD/olf@wo && \
  sleep 1 && \
  run_silent kill $pidc && \
  true
}

test_serve_xobs_obs() {
  info test_serve_xobs_obs && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  obs=xobs:${TD}/obs && \
  make_obs $BTD/obs $obs $TD/wc && \
  run_error cabri webapi --cdir $TD/wc obs+http://localhost:3000/$TD/obs@wo 2> /dev/null && \
  run_bg_silent cabri webapi --cdir $TD/wc xobs+http://localhost:3000/$TD/obs@wo && \
  sleep 1 && \
  run_silent kill $pidc && \
  true
}

test_basic_sync() {
  test_basic_sync_olf && \
  test_basic_sync_polf && \
  test_basic_sync_xolf && \
  test_basic_sync_obs && \
  test_basic_sync_xobs && \
  test_basic_sync_wolf_noe && \
  test_basic_sync_wolf_err && \
  test_basic_sync_wobs && \
  test_basic_sync_wobs_noe && \
  test_basic_sync_wobs_err && \
  test_basic_sync_xwolf && \
  test_basic_sync_xwobs && \
  test_basic_sync_wfs && \
  true
}

test_more_sync() {
  test_advanced_sync_olf && \
  test_advanced_sync_xolf && \
  test_sync_back_and_forth_olf && \
  test_sync_back_and_forth_xolf && \
  test_sync_back_and_forth_xwolf && \
  test_acl_sync_olf && \
  true
}

test_unlock() {
  test_basic_unlock_olf && \
  test_basic_unlock_xolf && \
  test_basic_unlock_obs && \
  test_basic_unlock_xobs && \
  test_basic_unlock_wolf && \
  test_basic_unlock_wobs && \
  test_basic_unlock_xwolf && \
  test_basic_unlock_xwobs && \
  true
}

test_index() {
  test_index_olf && \
  test_index_polf && \
  test_index_xolf && \
  test_index_obs && \
  test_index_xobs && \
  test_index_wolf && \
  test_index_wobs && \
  test_index_xwolf && \
  test_index_xwobs && \
  true
}

test_rest_api() {
  rurl="http://0.0.0.0:3000/wo/"
  info test_rest_api && \
  setup_test && \
  untar_simple && \
  fsy=fsy:${TD}/simple && \
  olf=olf:${TD}/olf && \
  make_polf $BTD/olf $olf $TD/wc && \
  run_bg_silent cabri webapi rest --cdir $TD/wc olf+http://localhost:3000/$TD/olf@wo && \
  sleep 1 && \
  run_silent curl -X POST -H "Content-Type: application/json" "${rurl}?mtime=2023-06-14T19:04:44Z&child=d1/&child=f1" && \
  run_silent curl -X GET ${rurl} && \
  find_out "[\"d1/\",\"f1\"]" &&  \
  run_silent curl -X GET "${rurl}?meta" && \
  find_out "\":[\"d1/\",\"f1\"]" &&  \
  run_silent curl -X PUT -H "Content-Type: application/octet-stream" "${rurl}f1?mtime=2023-06-14T19:05:45Z" --data-binary @$TD/simple/d1/f1 && \
  run_silent curl -X GET "${rurl}f1?meta" && \
  find_out "size\":3" &&  \
  run_silent curl -X GET "${rurl}f1" && \
  find_out "f1" && \
  run_silent curl -X DELETE "${rurl}f1" && \
  run_silent curl -X GET ${rurl} && \
  find_out "[\"d1/\"]" &&  \
  run_silent curl -X POST -H "Content-Type: application/json" "${rurl}?mtime=2023-06-14T19:04:44Z&child=d1/&child=f1sl" && \
  run_silent curl -X POST -H "Content-Type: application/json" "${rurl}f1sl?mtime=2023-06-14T19:04:45Z&child=d1/&symlink=d1/targetf1" && \
  run_silent curl -X GET "${rurl}f1sl?meta" && \
  find_out "ch\":\"7387f77968d069db425fd8761690877d" &&  \
  run_silent kill $pidc && \
  true
}

test_fixes() {
  test_serve_olf_xolf && \
  test_serve_obs_xobs && \
  test_serve_xolf_olf && \
  test_serve_xobs_obs && \
  true
}

PATH="$base_dir/build:$HOME/go/bin:$PATH"
OBS_ENV="--obsrg $OVHRG --obsep $OVHEP --obsct $OVHCT --obsak $OVHAK --obssk $OVHSK"
st=0
test_cli_fast=
info "starting"
true && \
  run_command cabri version && \
  test_basic_sync && \
  test_more_sync && \
  test_unlock && \
  test_index && \
  test_rest_api && \
  test_fixes && \
  true || (info failed && exit 1)
st=$?
info "ended"
exit $st

}