param(
  [string]$machine="manjaro",
  [string]$snapshot="ssh",
  [string]$script="run-full"
)

. "./env.ps1"

$is_mac=$machine.startswith("mac-")
$is_win=$machine.startswith("win-")
$is_lin=(!$is_win -and !$is_mac)

# # clean up
# ssh -o ConnectTimeout=1 -t -p 1111 zwhitchcox@localhost "zetup id delete --all"
# ssh-keygen -R [localhost]:1111

# # turn off current vm and restore snapshot
# VBoxManage controlvm "$machine" poweroff
# VBoxManage snapshot "$machine" restore $snapshot
# VBoxManage startvm "$machine" --type headless

# echo "Waiting for ssh connectivity"


# # get home dir
# if ($is_mac) {
#   $home_dir="/Users/zwhitchcox"
# } elseif ($is_win) {
#   $home_dir="C:/Users/zwhitchcox"
# } else {
#   $home_dir="/home/zwhitchcox"
# }

# # copy public key to vm
# cat "${HOME}\.ssh\id_rsa.pub" | ssh zwhitchcox@localhost -p 1111 "mkdir -p ~/.ssh && cat >> ${home_dir}/.ssh/authorized_keys"

# # copy zetup-pkg
# $pkg_dir="${home_dir}/.zetup/pkg/zetup-sh"
# ssh -p 1111 zwhitchcox@localhost "rm -rf ${pkg_dir}/* && mkdir -p ${pkg_dir}"
# scp -P 1111 -r ${HOME}/dev/zetup-pkg  "zwhitchcox@localhost:${pkg_dir}"
# ssh -p 1111 zwhitchcox@localhost "dos2unix --quiet ${pkg_dir}{*,/**/*}"


# # build zetup and copy to vm
# rm -r -Force ./build/*
# if ($is_mac) {
#   $platform="darwin"
# } elseif ($is_win) {
#   $platform="windows"
# } else {
#   $platform="linux"
# }
# sh ./scripts/cross-platform-build.sh "${platform}/amd64"
# scp -P 1111 -r "./build/zetup-${platform}-amd64"  "zwhitchcox@localhost:/tmp/zetup"

function run-tmp-script() {
  Param ([string]$scriptname, [string]$executionargs)
  $awk_replace="awk '{ sub(\`"\r$\`", \`"\`"); print }'"
  scp -P 1111 -r "${scriptname}"  "zwhitchcox@localhost:/tmp/script.sh"
  ssh -p 1111 zwhitchcox@localhost "$awk_replace /tmp/script.sh > /tmp/script_unix.sh"
  ssh -p 1111 zwhitchcox@localhost "bash /tmp/script_unix.sh $executionargs"
}

# move zetup
run-tmp-script "./tests/mv-zetup.sh" "${VM_PASS}"

# run zetup on vm
# run-tmp-script "./tests/run-zetup-pkg.sh" "${GITHUB_PASS}"