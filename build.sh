#!/bin/sh
# Shell script to build GNU Make in the absence of any 'make' program.
# build.sh.  Generated from build.sh.in by configure.

# Copyright (C) 1993-2016 Free Software Foundation, Inc.
# This file is part of GNU Make.
#
# GNU Make is free software; you can redistribute it and/or modify it under
# the terms of the GNU General Public License as published by the Free Software
# Foundation; either version 3 of the License, or (at your option) any later
# version.
#
# GNU Make is distributed in the hope that it will be useful, but WITHOUT ANY
# WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
# FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
# details.
#
# You should have received a copy of the GNU General Public License along with
# this program.  If not, see <http://www.gnu.org/licenses/>.

# See Makefile.in for comments describing these variables.

srcdir='/Users/akutz/.opt/make/4.2//make-4.2'
CC='gcc'
CFLAGS='-g -O2 '
CPPFLAGS=''
LDFLAGS='-rdynamic '
ALLOCA=''
LOADLIBES='  -lintl -Wl,-framework -Wl,CoreFoundation'
eval extras=\'\'
REMOTE='stub'
GLOBLIB='glob/libglob.a'
PATH_SEPARATOR=':'
OBJEXT='o'
EXEEXT=''

# Common prefix for machine-independent installed files.
prefix='/Users/akutz/.opt/make/4.2'
# Common prefix for machine-dependent installed files.
exec_prefix=`eval echo ${prefix}`
# Directory to find libraries in for '-lXXX'.
libdir=${exec_prefix}/lib
# Directory to search by default for included makefiles.
includedir=${prefix}/include

localedir=${prefix}/share/locale
aliaspath=${localedir}${PATH_SEPARATOR}.

defines="-DLOCALEDIR=\"${localedir}\" -DLIBDIR=\"${libdir}\" -DINCLUDEDIR=\"${includedir}\""' -DHAVE_CONFIG_H'

# Exit as soon as any command fails.
set -e

# These are all the objects we need to link together.
objs="ar.${OBJEXT} arscan.${OBJEXT} commands.${OBJEXT} default.${OBJEXT} dir.${OBJEXT} expand.${OBJEXT} file.${OBJEXT} function.${OBJEXT} getopt.${OBJEXT} getopt1.${OBJEXT} guile.${OBJEXT} implicit.${OBJEXT} job.${OBJEXT} load.${OBJEXT} loadapi.${OBJEXT} main.${OBJEXT} misc.${OBJEXT} posixos.${OBJEXT} output.${OBJEXT} read.${OBJEXT} remake.${OBJEXT} rule.${OBJEXT} signame.${OBJEXT} strcache.${OBJEXT} variable.${OBJEXT} version.${OBJEXT} vpath.${OBJEXT} hash.${OBJEXT} remote-${REMOTE}.${OBJEXT} ${extras} ${ALLOCA}"

if [ x"$GLOBLIB" != x ]; then
  objs="$objs glob/fnmatch.${OBJEXT} glob/glob.${OBJEXT}"
  globinc=-I${srcdir}/glob
fi

# Compile the source files into those objects.
for file in `echo ${objs} | sed 's/\.'${OBJEXT}'/.c/g'`; do
  echo compiling ${file}...
  $CC $defines $CPPFLAGS $CFLAGS \
      -c -I. -I${srcdir} ${globinc} ${srcdir}/$file
done

# The object files were actually all put in the current directory.
# Remove the source directory names from the list.
srcobjs="$objs"
objs=
for obj in $srcobjs; do
  objs="$objs `basename $obj`"
done

# Link all the objects together.
echo linking make...
$CC $CFLAGS $LDFLAGS $objs $LOADLIBES -o makenew${EXEEXT}
echo done
mv -f makenew${EXEEXT} make${EXEEXT}
