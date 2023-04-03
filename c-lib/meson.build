# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2022-2024 Intel Corporation

project('go-pktgen', 'c',
    version: run_command(find_program('cat', 'more'),
        files('../VERSION'), check: false).stdout().strip(),

    license: 'BSD',
    default_options: [
        'buildtype=release',
        'default_library=static',
        'warning_level=3',
        'werror=true'
    ],
    meson_version: '>= 0.61.2'
)

use_static_libs = get_option('default_library') == 'static'

gpkt_libs = []
enabled_libs = []
extra_ldflags = []

gpkt_conf = configuration_data()

machine_args = []

# set up some global vars for compiler, platform, configuration, etc.
cc = meson.get_compiler('c')

add_project_arguments('-D__PROJECT_VERSION="' + meson.project_version() + '"', language: 'c')

# specify -D_GNU_SOURCE unconditionally
add_project_arguments('-D_GNU_SOURCE', language: 'c')

# Add the gpkt_build_config.h as a default include for all files.
global_inc = include_directories('.')
build_cfg_file = 'gpkt_build_config.h'
add_project_arguments(['-include', build_cfg_file] + machine_args, language: 'c')
add_project_arguments(['-include', build_cfg_file] + machine_args, language: 'cpp')

machine = get_option('machine')

# machine type 'default' is special, it defaults to the per arch agreed common
# minimal baseline needed for CNDP.
# That might not be the most optimized, but the most portable version while
# still being able to support the CPU features required for CNDP.
# This can be bumped up by the CNDP project, but it can never be an
# invariant like 'native'
if machine == 'default'
    if host_machine.cpu_family().startswith('x86')
        # matches the old pre-meson build systems default
        machine = 'corei7'
    endif
endif

gpkt_conf.set('GPKT_MACHINE', machine)
gpkt_conf.set('GPKT_CACHE_LINE_SIZE', 64)

machine_args += '-march=' + machine

toolchain = cc.get_id()
gpkt_conf.set_quoted('GPKT_TOOLCHAIN', toolchain)
gpkt_conf.set('GPKT_TOOLCHAIN_' + toolchain.to_upper(), 1)

gpkt_conf.set('GPKT_ARCH_64', cc.sizeof('void *') == 8)

add_project_link_arguments('-Wl,--no-as-needed', language: 'c')
if get_option('buildtype') == 'release'
    add_project_arguments('-g', language: 'c')
endif

# use pthreads
add_project_link_arguments('-pthread', language: 'c')
extra_ldflags += '-pthread'

add_project_link_arguments('-lm', language: 'c')
extra_ldflags += '-lm'

# on Alpine, execinfo and backtrace are in a separate library
if not cc.has_header('execinfo.h')
    error('Missing execinfo.h. Try installing "libexecinfo-dev".')
endif

if dependency('libexecinfo', required : false, static : use_static_libs).found()
    add_project_link_arguments('-lexecinfo', language: 'c')
    extra_ldflags += '-lexecinfo'
endif

# check for libraries used in multiple places in CNDP
has_libnuma = 0
numa_dep = dependency('numa', required: false, static: use_static_libs)
if numa_dep.found() and cc.has_header('numaif.h')
    gpkt_conf.set10('GPKT_HAS_LIBNUMA', true)
    has_libnuma = 1
    add_project_link_arguments('-lnuma', language: 'c')
    extra_ldflags += '-lnuma'
endif

# check for libbsd
libbsd = dependency('libbsd', required: true, static: use_static_libs)
if libbsd.found()
    add_project_link_arguments('-lbsd', language: 'c')
    gpkt_conf.set('GPKT_USE_LIBBSD', 1)
endif

# if link_lib is empty, do not add it to project properties
add_project_link_arguments('-ldl', language: 'c')
extra_ldflags += '-ldl'

dpdk = dependency('libdpdk', required: true)
message('prefix: ' + get_option('prefix') + ' libdir: ' + get_option('libdir'))

dpdk_libs_path = join_paths(get_option('prefix'), get_option('libdir'))
message('DPDK lib path: ' + dpdk_libs_path)

dpdk_cflags = ['-DALLOW_EXPERIMENTAL_API', '-DALLOW_INTERNAL_API']

# enable extra warnings and disable any unwanted warnings
warning_flags = [
    '-Wno-pedantic',
    '-Wno-format-truncation',
]
foreach arg: warning_flags
    if cc.has_argument(arg)
        add_project_arguments(arg, language: 'c')
    endif
endforeach

# write the build config into the cndp include directory
configure_file(output: build_cfg_file,
    configuration: gpkt_conf,
    install_dir: join_paths(get_option('includedir'), meson.project_name().to_lower()))

build_cfg = declare_dependency(include_directories: include_directories('.'))

subdir('gpkt')

libgpkt_a = []
libgpkt_so = []
foreach lib:enabled_libs
    libgpkt_a += 'lib' + lib + '.a '
    libgpkt_so += 'lib' + lib + '.so '
endforeach

gpkt_a_name = 'libgpkt.a'
gpkt_so_name = 'libgpkt.so'

mklib = find_program('scripts/mklib.sh')
build_dir = meson.current_build_dir()

if use_static_libs
    gpkt_a = custom_target('libgpkt_a_target',
        output: 'libgpkt.a',
        command:[mklib, build_dir, '@OUTPUT@', libgpkt_a],
        install_dir: join_paths('lib', 'x86_64-linux-gnu'),
        install: true)
else
    gpkt_so = custom_target('libgpkt_so_target',
        output: 'libgpkt.so',
        command:[mklib, build_dir, '@OUTPUT@', libgpkt_so],
        install_dir: join_paths('lib', 'x86_64-linux-gnu'),
        install: true)
endif

pkg = import('pkgconfig')

message('>>> Create pkg-config file')
pkg.generate(name: meson.project_name(),
    filebase: 'lib' + meson.project_name().to_lower(),
    version: meson.project_version(),
    libraries: ['-Wl,--as-needed'] + gpkt_libs,
    libraries_private: ['-lgpkt'] + ['-Wl,-Bdynamic'] + extra_ldflags,
    requires: libbsd, # if libbsd is not enabled, then this is blank
    description: '''The Go-Pktgen traffic generator (Go-Pktgen).
Note that CFLAGS might contain an -march flag higher than typical baseline.
This is required for a number of static inline functions in the public headers.''',
    subdirs: ['gpkt'],
    extra_cflags: ['-include', build_cfg_file] + machine_args,
    install_dir: 'lib/pkgconfig'
)
message('<<< Done pkg-config file')
