#!/bin/bash
# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2023-2025 Intel Corporation

# A simple script to help build Go-Pktgen using meson/ninja tools.
# The script also creates an installed directory called usr/local.
# The install directory will contain all of the includes and libraries
# for external applications to build and link with Go-Pktgen.
#
# using 'gpkt-build.sh help' or 'gpkt-build.sh -h' or 'gpkt-build.sh --help' to see help information.
#

currdir=$(pwd)
script_dir=$(cd "${BASH_SOURCE[0]%/*}" && pwd -P)
sdk_dir="${GPKT_SDK_DIR:-${script_dir%/*}}"
target_dir="${GPKT_TARGET_DIR:-usr/local}"
build_dir="${GPKT_BUILD_DIR:-${currdir}/_build}"
install_path="${GPKT_DEST_DIR:-${currdir}}"

export PKG_CONFIG_PATH="${PKG_CONFIG_PATH:-/usr/lib64/pkgconfig}"

buildtype="release"
static=""
configure="setup"

if [[ "${build_dir}" = /* ]]; then
    # absolute path to build dir. Don't prepend workdir.
    build_path=$build_dir
else
    build_path=${currdir}/$build_dir
fi
if [[ ! "${install_path}" = /* ]]; then
    # relative path for install path detected
    # prepend with currdir
    install_path=${currdir}/${install_path}
fi
if [[ "${target_dir}" = .* ]]; then
    echo "target_dir starts with . or .. if different install prefix required then use GPKT_DEST_DIR instead";
    exit 1;
fi
if [[ "${target_dir}" = /* ]]; then
    echo "target_dir absolute path detected removing leading '/'"
    export target_dir=${target_dir##/}
fi
target_path=${install_path%/}/${target_dir%/}

echo "Build environment variables and paths:"
echo "  GPKT_SDK_DIR    : $sdk_dir"
echo "  GPKT_TARGET_DIR : $target_dir"
echo "  GPKT_BUILD_DIR  : $build_dir"
echo "  GPKT_DEST_DIR   : $install_path"
echo "  PKG_CONFIG_PATH : $PKG_CONFIG_PATH"
echo "  build_path      : $build_path"
echo "  target_path     : $target_path"
echo "  install_path    : $install_path"
echo ""

function run_meson() {
    btype="-Dbuildtype=$buildtype"
    if [[ $verbose = true ]]; then
        echo "*** meson $configure $static $btype --prefix=/$target_dir $build_path $sdk_dir"
    fi
    if ! meson $configure $static $btype --prefix="/$target_dir" "$build_path" "$sdk_dir"; then
        echo "*** ERROR: meson $configure $static $btype --prefix=/$target_dir $build_path $sdk_dir"
        configure=""
        return 1;
    fi
    if [[ $verbose = true ]]; then
        echo "*** Done running meson"
    fi
    configure=""

    return 0
}

function ninja_build() {
    echo ">>> Ninja build in $build_path buildtype= $buildtype"

    if [[ -d $build_path ]] || [[ -f $build_path/build.ninja ]]; then
        # add reconfigure command if meson dir already exists
        configure="configure"
        # sdk_dir must be empty if we're reconfiguring
        sdk_dir=""
    fi
    if ! run_meson; then
        return 1;
    fi

    if ! ninja -C "$build_path"; then
        return 1;
    fi

    return 0
}

function ninja_build_docs() {
    echo ">>> Ninja build documents in $build_path"

    if [[ ! -d $build_path ]] || [[ ! -f $build_path/build.ninja ]]; then
        if ! run_meson; then
            return 1;
        fi
    fi

    if ! ninja -C "$build_path" doc; then
        return 1;
    fi
    return 0
}

ninja_install() {
    echo ">>> Ninja install to $target_path"

    if [[ $verbose = true ]]; then
        if ! DESTDIR=$install_path ninja -C "$build_path" install; then
            echo "*** Install failed!!"
            return 1
        fi
    else
        if ! DESTDIR=$install_path ninja -C "$build_path" install > /dev/null; then
            echo "*** Install failed!!"
            return 1
        fi
    fi

    return 0
}

ninja_uninstall() {
    echo ">>> Ninja uninstall to $target_path"

    if [[ $verbose = true ]]; then
        if ! DESTDIR=$install_path ninja -C "$build_path" uninstall; then
            echo "*** Uninstall failed!!"
            return 1;
        fi
    else
        if ! DESTDIR=$install_path ninja -C "$build_path" uninstall > /dev/null; then
            echo "*** Uninstall failed!!"
            return 1;
        fi
    fi

    return 0
}

usage() {
    echo " Usage: Build Go-Pktgen using Meson/Ninja tools"
    echo "  ** Must be in the top level directory for Go-Pktgen"
    echo "     This tool is in tools/gpkt-build.sh, but use 'make' which calls this script"
    echo "     Use 'make' to build Go-Pktgen as it allows for multiple targets i.e. 'make clean debug'"
    echo ""
    echo "     GPKT_SDK_DIR    - Go-Pktgen source directory path (default: current working directory)"
    echo "     GPKT_TARGET_DIR - Target directory for installed files (default: usr/local)"
    echo "     GPKT_BUILD_DIR  - Build directory name (default: _build)"
    echo "     GPKT_DEST_DIR   - Destination directory (default: current working directory)"
    echo ""
    echo "  gpkt-build.sh    - create the 'build_dir' directory if not present and compile Go-Pktgen"
    echo "                     If the 'build_dir' directory exists it will use ninja to build Go-Pktgen"
    echo "                     without running meson unless one of the meson.build files were changed"
    echo "    -v             - Enable verbose output"
    echo "    build          - build Go-Pktgen using the 'build_dir' directory"
    echo "    static         - build Go-Pktgen static using the 'build_dir' directory, 'make static build'"
    echo "    debug          - turn off optimization, may need to do 'clean' then 'debug' the first time"
    echo "    debugopt       - turn optimization on with -O2, may need to do 'clean' then 'debugopt'"
    echo "                     the first time"
    echo "    clean          - remove the 'build_dir' directory then exit"
    echo "    install        - install the includes/libraries into 'target_dir' directory"
    echo "    uninstall      - uninstall the includes/libraries from 'target_dir' directory"
    echo "    docs           - create the document files"
    exit
}

verbose=false

echo "==== command: $@"
for cmd in "$@"
do
    case "$cmd" in
    'help' | '-h' | '--help')
        usage
        ;;

    '-v' | '--verbose')
        verbose=true
        ;;

    'static')
        echo ">>> Static  build in $build_path"
        static="-Ddefault_library=static"
        ;;

    'build')
        echo ">>> Release build in $build_path"
        ninja_build
        ;;

    'debug')
        echo ">>> Debug build in $build_path"
        buildtype="debug"
        ninja_build
        ;;

    'debugopt')
        echo ">>> Debug Optimized build in $build_path"
        buildtype="debugoptimized"
        ninja_build
        ;;

    'clean')
    echo "*** Removing $build_path directory"
        rm -fr "$build_path"
        ;;

    'uninstall')
        echo "*** Uninstalling $target_path directory"
        ninja_uninstall
        exit
        ;;

    'install')
        echo ">>> Install the includes/libraries into $target_path directory"
        ninja_install
        ;;

    'docs')
        echo ">>> Create the documents in $build_path directory"
        ninja_build_docs
        ;;

    *)
        if [[ $# -gt 0 ]]; then
            usage
        else
            echo ">>> Build and install Go-Pktgen"
            ninja_build && ninja_install
        fi
        ;;
    esac
done
