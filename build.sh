#!/bin/sh

################################################################################
# build REX-Ray
#
# usage: build.sh [-b builder] [-x] [-t type] [-d drivers] [FILE]
#
#       -b builder       Optional. The builder used to build REX-Ray. Possible
#                        values include docker and make. The default value, if
#                        installed and running, is docker. If Docker is not
#                        installed or the current user cannot access it then
#                        make is used.
#
#                        If docker is *explicitly* specified but not installed
#                        or inaccessible, this program will *not* switch to
#                        make, but instead exit with an error.
#
#       -x               Optional. This flag is only applicable to the make
#                        build runner. If set, this flag prevents make from
#                        performing a clean ahead of the build. Therefore
#                        this flag results in the preservation of e-X-isting
#                        files.
#
#                        This flag sets "-b make".
#
#       -t type          Optional. The type of REX-Ray binary to create.
#                        Possible values include: agent, client, and controller
#                        The default value is the same as omitting this flag and
#                        argument altogether and will create a binary that
#                        includes the agent, client, and controller.
#
#       -d drivers       Optional. One or more drivers to include in the binary.
#                        Specify multiple drivers as a quoted, space-delimited
#                        list.
#
#                        Drivers are only included in standard and controller
#                        binaries. Please see the -t type flag for more
#                        information on binary types.
#
#       -u uri           Optional. The URI to a REX-Ray, git repository . This
#                        flag is only valid for the docker builder. Specifying
#                        a repository will use its sources (specified by the
#                        associated -r ref flag) instead of the local sources.
#
#       -r ref           Optional. The git references to use when specifying the
#                        -u uri flag. This value can be a commit ID, tag, or
#                        branch name. The default value is master.
#
#                        This flag sets "-b docker".
#
#                        If -r is set but -u is not, -u will be set to
#                        to libStorage's primary repository URI.
#
#       -l               Optional. A flag indicating to use the local libStorage
#                        sources instead of the libStorage version specified in
#                        REX-Ray's glide.yaml file.
#
#                        This flag sets "-b docker" and cannot be used with the
#                        -1 or -2 flags.
#
#       -1 uri           Optional. The URI to a libStorage, git repository. This
#                        flag is only valid for the docker builder. Specifying
#                        a repository will use its sources (specified by the
#                        associated -2 ref flag) instead of the libStorage
#                        version specified in REX-Ray's glide.yaml file.
#
#                        This flag sets "-b docker" and cannot be used with the
#                        -l flag.
#
#       -2 ref           Optional. The git references to use when specifying the
#                        -1 uri flag. This value can be a commit ID, tag, or
#                        branch name. The default value is master.
#
#                        This flag sets "-b docker" and cannot be used with the
#                        -l flag.
#
#                        If -2 is set but -1 is not, -1 will be set to
#                        to libStorage's primary repository URI.
#
#       FILE             Optional. The name of the produced binary. The default
#                        name of the binary is based on the binary type. For
#                        example, if -t client is set then the file will be
#                        rexray-client. If no type is set then the default file
#                        name is rexray.
#
################################################################################

# the version of go to use
GO_VERSION="1.8.1"

# the makeflags to use for make commands
MAKEFLAGS=--no-print-directory

usage() {
  echo "usage: build.sh [-b builder] [-x] [-t type] [-d drivers]"
  echo "                [-u rr_uri] [-r rr_ref]"
  echo "                [-l]"
  echo "                [-1 ls_uri] [-2 ls_ref]"
  echo "                [FILE]"
  echo
  echo "       -b builder       Optional. The builder used to build REX-Ray. Possible"
  echo "                        values include docker and make. The default value, if"
  echo "                        installed and running, is docker. If Docker is not"
  echo "                        installed or the current user cannot access it then"
  echo "                        make is used."
  echo
  echo "                        If docker is *explicitly* specified but not installed"
  echo "                        or inaccessible, this program will *not* switch to"
  echo "                        make, but instead exit with an error."
  echo
  echo "       -x               Optional. This flag is only applicable to the make"
  echo "                        build runner. If set, this flag prevents make from"
  echo "                        performing a clean ahead of the build. Therefore"
  echo "                        this flag results in the preservation of e-X-isting"
  echo "                        files."
  echo
  echo "                        This flag sets '-b make'."
  echo
  echo "       -t type          Optional. The type of REX-Ray binary to create."
  echo "                        Possible values include: agent, client, and controller"
  echo "                        The default value is the same as omitting this flag and"
  echo "                        argument altogether and will create a binary that"
  echo "                        includes the agent, client, and controller."
  echo
  echo "       -d drivers       Optional. One or more drivers to include in the binary."
  echo "                        Specify multiple drivers as a quoted, space-delimited"
  echo "                        list."
  echo
  echo "                        Drivers are only included in standard and controller"
  echo "                        binaries. Please see the -t type flag for more"
  echo "                        information on binary types."
  echo
  echo "       -u uri           Optional. The URI to a REX-Ray, git repository . This"
  echo "                        flag is only valid for the docker builder. Specifying"
  echo "                        a repository will use its sources (specified by the"
  echo "                        associated -r ref flag) instead of the local sources."
  echo
  echo "       -r ref           Optional. The git references to use when specifying the"
  echo "                        -u uri flag. This value can be a commit ID, tag, or"
  echo "                        branch name. The default value is master."
  echo
  echo "                        This flag sets '-b docker'."
  echo
  echo "                        If -r is set but -u is not, -u will be set to"
  echo "                        to libStorage's primary repository URI."
  echo
  echo "       -l               Optional. A flag indicating to use the local libStorage"
  echo "                        sources instead of the libStorage version specified in"
  echo "                        REX-Ray's glide.yaml file."
  echo
  echo "                        This flag sets '-b docker' and cannot be used with the"
  echo "                        -1 or -2 flags."
  echo
  echo "       -1 uri           Optional. The URI to a libStorage, git repository. This"
  echo "                        flag is only valid for the docker builder. Specifying"
  echo "                        a repository will use its sources (specified by the"
  echo "                        associated -2 ref flag) instead of the libStorage"
  echo "                        version specified in REX-Ray's glide.yaml file."
  echo
  echo "                        This flag sets '-b docker' and cannot be used with the"
  echo "                        -l flag."
  echo
  echo "       -2 ref           Optional. The git references to use when specifying the"
  echo "                        -1 uri flag. This value can be a commit ID, tag, or"
  echo "                        branch name. The default value is master."
  echo
  echo "                        This flag sets '-b docker' and cannot be used with the"
  echo "                        -l flag."
  echo
  echo "                        If -2 is set but -1 is not, -1 will be set to"
  echo "                        to libStorage's primary repository URI."
  echo
  echo "       FILE             Optional. The name of the produced binary. The default"
  echo "                        name of the binary is based on the binary type. For"
  echo "                        example, if -t client is set then the file will be"
  echo "                        rexray-client. If no type is set then the default file"
  echo "                        name is rexray."
  exit 1
}

# get_git_repo returns a github url.
#
#   $1    the slug used to seed the url
#   $2    optional. a git repo name
#
# if the slug starts with http|https|ssh|git then
# the slug is returned as is.
#
# if the slug contains a / character, then the slug
# is appended to https://github.com/ and returned
#
# if the $2 parameter is not defined the slug is used
# to build https://github.com/${1}/rexray and returned
#
# the slug and $2 parameter are used to build
# https://github.com/${1}/${2} and returned
get_git_repo() {
  if [ "$1" = "" ]; then
    return 0
  fi
  if echo "$1" | grep -i '^\(http\|https\|ssh\|git\)' > /dev/null 2>&1; then
    echo "$1"
    return 0
  fi
  if echo "$1" | grep '/' > /dev/null 2>&1; then
    echo "https://github.com/${1}"
    return 0
  fi
  if [ "$2" = "" ]; then
    echo "https://github.com/${1}/rexray"
    return 0
  else
    echo "https://github.com/${1}/${2}"
    return 0
  fi
  return 1
}

# git_clone_checkout_cmd returns the commands to git a repository,
# checkout a ref, and creating a branch for a detached state if
# necessary
#
#     $1    the rexray git repo
#     $2    the rexray git ref
git_clone_checkout_cmds() {
  uri="$(get_git_repo "$1")"
  ref="${2:-master}"

  cmd=' git clone -q '"$uri"' . > /dev/null && '
  cmd="$cmd"'git checkout -q '"$ref"' > /dev/null && '
  cmd="$cmd"'if git status | grep "HEAD detached" > /dev/null 2>&1; then '
  cmd="$cmd"'git checkout -q -b '"$ref"' > /dev/null; fi'
  echo "$cmd"
  return 0
}

# get_semver returns the project's semantic version
#
#     $1    the builder type
#     $2    optional. the rexray git repo
#     $3    optional. the rexray git ref
get_semver() {
  builder="${1:-docker}"

  if [ "$builder" = "make" ]; then
    version="$(PORCELAIN=1 make version-porcelain)"
    if [ "$?" -ne "0" ]; then return 1; fi
    echo "$version" | tr -d "\r" | tr -d "\n"
    return 0
  fi

  rr_uri="$(get_git_repo "$2")"
  rr_ref="${3:-master}"
  go_ver="$GO_VERSION"

  vcmd="PORCELAIN=1 make $MAKEFLAGS -C /rexray version-porcelain"
  if [ "$rr_uri" = "" ]; then
    version=$(docker run -it --rm -v "$(pwd)":/rexray \
      golang:${go_ver} bash -c ''"$vcmd"'')
    if [ "$?" -ne "0" ]; then return 1; fi
    echo "$version" | tr -d "\r" | tr -d "\n"
    return 0
  fi

  mkgd=" mkdir -p /rexray && cd /rexray"
  vcmd="$mkgd && $(git_clone_checkout_cmds "$rr_uri" "$rr_ref") && $vcmd"
  version=$(docker run -it --rm golang:${go_ver} bash -c ''"$vcmd"'')
  if [ "$?" -ne "0" ]; then return 1; fi
  echo "$version" | tr -d "\r" | tr -d "\n"
  return 0
}

create_dockerfile() {
  workdir_rr="/go/src/github.com/codedellemc/rexray/"
  bsrc='NOSTAT=1 DRIVERS="'"${DRVRS}"'" REXRAY_BUILD_TYPE='"${BTYPE}"' make'

  lsd="${GOPATH}/src/github.com/codedellemc/libstorage"
  if [ "$LS_LOCAL" = "1" ] && [ -d "$lsd" ]; then
    tar -czf ".ls.tar.gz" \
      --exclude "./.git" \
      --exclude "./vendor" \
      --exclude "*.test" \
      -C "$lsd" .

    ls_home="/go/src/github.com/codedellemc/libstorage/"
    workdir_ls="WORKDIR $ls_home"

    nl="$(printf '%b_' '\n')"
    sed -e 's/.*# libstorage-version/    ref:     master/g' \
        -e 's|.*# libstorage-repo|    repo:    file://'"${ls_home}\\${nl%_}"'    vcs:     git|g' \
        glide.yaml > .glide.yaml.tmp
    bsrc="cp -f .glide.yaml.tmp glide.yaml \&\& $bsrc"

    init_ls_srcs="RUN tar xzf ${workdir_rr}.ls.tar.gz \&\& "
    init_ls_srcs="${init_ls_srcs}git init \&\& "
    init_ls_srcs="${init_ls_srcs}git config --local user.name $(whoami) \&\& "
    init_ls_srcs="${init_ls_srcs}git config --local user.email $(whoami)@localhost \&\& "
    init_ls_srcs="${init_ls_srcs}git add -A > /dev/null \&\& "
    init_ls_srcs="${init_ls_srcs}git commit -m v0.0.1 > /dev/null \&\& "
    init_ls_srcs="${init_ls_srcs}git tag -a -m v0.0.1 v0.0.1"

  elif [ ! "$LS_URI" = "" ]; then
     nl="$(printf '%b_' '\n')"
     sed -e 's/.*# libstorage-version/    ref:     '"$LS_REF"'/g' \
         -e 's|.*# libstorage-repo|    repo:    '"$LS_URI\\${nl%_}"'    vcs:     git|g' \
         glide.yaml > .glide.yaml.tmp
     bsrc="cp -f .glide.yaml.tmp glide.yaml \&\& $bsrc"
  fi

  if [ "$RR_URI" = "" ]; then
    copy_rr_srcs='COPY . .'
    bcmd="$bsrc"
  else
    if [ "$LS_LOCAL" = "1" ]; then
      copy_rr_srcs='COPY [".ls.tar.gz", ".glide.yaml.tmp", "./"]'
    elif [ ! "$LS_URI" = "" ]; then
      copy_rr_srcs='COPY .glide.yaml.tmp .'
    fi
    bgit='(mv .glide.yaml.tmp /tmp/ > /dev/null 2>\&1 || true)'
    bgit="$bgit"' \&\& git clone '"$RR_URI"' . \&\& git checkout '"$RR_REF"
    bgit="$bgit"' \&\& if git status | grep "HEAD detached" > /dev/null; '
    bgit="$bgit"'then git checkout -b '"$RR_REF"'; fi'
    bgit="$bgit"' \&\& (mv /tmp/.glide.yaml.tmp . > /dev/null 2>\&1 || true)'
    bcmd="$bgit"' \&\& '"$bsrc"
  fi

  sed -e 's/@GO_VERSION@/'"$GO_VERSION"'/g' \
    -e 's|@WORKDIR_RR@|'"$workdir_rr"'|g' \
    -e 's|@WORKDIR_LS@|'"$workdir_ls"'|g' \
    -e 's|@INIT_LS_SRCS_CMD@|'"$init_ls_srcs"'|g' \
    -e 's|@COPY_RR_SRCS_CMD@|'"$copy_rr_srcs"'|g' \
    -e 's%@BUILD_CMD@%'"$bcmd"'%g' \
    -e 's/@BUILD_TYPE@/'"$BTYPE"'/g' \
    -e 's/@FNAME_SUFFIX@/'"$FNAME_SUFFIX"'/g' \
    -e 's|@SEMVER@|'"$SEMVER"'|g' \
    -e 's|@DRIVERS@|'"$DRVRS"'|g' \
    -e 's/@DOCKERFILE@/'"$DOCKERFILE_TMP"'/g' \
    "$DOCKERFILE_SRC" > "$DOCKERFILE_TMP"

  return 0
}

# the builder
BUILDER="docker"

# the build type
BTYPE="${REXRAY_BUILD_TYPE:-}"

# the drivers
DRVRS="${DRIVERS:-}"

# the file name
FNAME=

# do not clean ahead of a make
NOCLN=

while getopts ":b:t:d:xu:r:l1:2:" opt; do
  case $opt in
  b)
    FLAG_B="1"
    BUILDER="$OPTARG"
    ;;
  t)
    FLAG_T="1"
    BTYPE="$OPTARG"
    ;;
  d)
    FLAG_D="1"
    DRVRS="$OPTARG"
    ;;
  x)
    FLAG_X="1"
    NOCLN="1"
    ;;
  u)
    FLAG_U="1"
    RR_URI="$OPTARG"
    ;;
  r)
    FLAG_R="1"
    RR_REF="$OPTARG"
    ;;
  l)
    FLAG_L="1"
    LS_LOCAL="1"
    ;;
  1)
    FLAG_1="1"
    LS_URI="$OPTARG"
    ;;
  2)
    FLAG_2="1"
    LS_REF="$OPTARG"
    ;;
  *)
    usage
    ;;
  esac
done
shift $((OPTIND-1))

# validate the builder
if [ ! "$BUILDER" = "docker" ] && [ ! "$BUILDER" = "make" ]; then
  echo "error: invalid builder: $BUILDER"
  echo
  usage
fi

# if the docker builder was selected then make sure it is available
if [ "$BUILDER" = "docker" ]; then
  if ! docker version > /dev/null 2>&1; then
    if [ "$FLAG_B" = "1" ]; then
      echo "error: docker builder unavailable"
      echo
      usage
    fi
    BUILDER="make"
  fi
fi

# validate the libstorage source location
if [ "$FLAG_L" = "1" ] && ([ "$FLAG_1" = "1" ] || [ "$FLAG_2" = "1" ]); then
  echo "error: cannot use both local & remote libStorage sources"
  echo
  usage
fi

# validate that no make-only flags were specified with a docker builder
if [ "$BUILDER" = "docker" ] && [ "$FLAG_X" = "1" ]; then
  echo "error: -x cannot be used with the docker builder"
  echo
  usage
fi

# validate that no docker-only flags were specified with a make builder
if [ "$BUILDER" = "make" ] && \
   ([ "$FLAG_U" = "1" ] || [ "$FLAG_R" = "1" ] || \
    [ "$FLAG_L" = "1" ] || [ "$FLAG_1" = "1" ] || [ "$FLAG_2" = "1" ]); then
  echo "error: -u,-r,-l,-1,-2 cannot be used with the make builder"
  echo
  usage
fi

# validate the build type
if [ ! "$BTYPE" = "" ] && \
   [ ! "$BTYPE" = "agent" ] && \
   [ ! "$BTYPE" = "client" ] && \
   [ ! "$BTYPE" = "controller" ]; then
   echo "error: invalid build type: $BTYPE"
   echo
   usage
fi

# if the -u flag was set then sanitize the uri
if [ "$FLAG_U" = "1" ]; then RR_URI="$(get_git_repo $RR_URI rexray)"; fi

# if the -1 flag was set then sanitize the uri
if [ "$FLAG_1" = "1" ]; then LS_URI="$(get_git_repo $LS_URI libstorage)"; fi

# if there is a rex-ray uri set, ensure the ref is defined
if [ ! "$RR_URI" = "" ]; then
  RR_REF="${RR_REF:-master}"
# if there is a rex-ray ref set, ensure the uri is defined
elif [ ! "$RR_REF" = "" ]; then
  RR_URI="https://github.com/codedellemc/rexray"
fi

# if there is a libstorage uri set, ensure the ref is defined
if [ ! "$LS_URI" = "" ]; then
  LS_REF="${LS_REF:-master}"
# if there is a libstorage ref set, ensure the uri is defined
elif [ ! "$LS_REF" = "" ]; then
  LS_URI="https://github.com/codedellemc/libstorage"
fi

# validate that drivers aren't set for agent or client builds
if [ ! "$DRVRS" = "" ] && \
  ([ "$BTYPE" = "agent" ] || [ "$BTYPE" = "client" ]); then
  echo "error: drivers are invalid for agent & client builds"
  echo
  usage
fi

if [ "$BTYPE" = "" ]; then
  FNAME="rexray"
  REAL_FNAME="rexray"
else
  FNAME="rexray-${BTYPE}"
  REAL_FNAME="rexray-${BTYPE}"
  FNAME_SUFFIX="-${BTYPE}"
fi

if [ ! "$1" = "" ]; then
  FNAME="$1"
fi

if [ "$BTYPE" = "" ]; then
  TAG="rexray"
  BTYPE="client+agent+controller"
else
  TAG="$BTYPE"
fi

if [ "$DEBUG" = "1" ]; then
  echo "FLAG_B=$FLAG_B"
  echo "FLAG_T=$FLAG_T"
  echo "FLAG_D=$FLAG_D"
  echo "FLAG_X=$FLAG_X"
  echo "FLAG_U=$FLAG_U"
  echo "FLAG_R=$FLAG_R"
  echo "FLAG_L=$FLAG_L"
  echo "FLAG_1=$FLAG_1"
  echo "FLAG_2=$FLAG_2"
  echo "TAG=$TAG"
  echo "NOCLN=$NOCLN"
  echo "BUILDER=$BUILDER"
  echo "RR_URI=$RR_URI"
  echo "RR_REF=$RR_REF"
  echo "LS_URI=$LS_URI"
  echo "LS_REF=$LS_REF"
  echo "LS_LOCAL=$LS_LOCAL"
  echo "BTYPE=$BTYPE"
  echo "DRVRS=$DRVRS"
  echo "FNAME=$FNAME"
  echo "REAL_FNAME=$REAL_FNAME"
  echo "FNAME_SUFFIX=$FNAME_SUFFIX"
fi

SEMVER="${SEMVER:-$(get_semver "$BUILDER" "$RR_URI" "$RR_REF")}"

if [ "$BUILDER" = "docker" ]; then
  DOCKERFILE_SRC=".Dockerfile"
  DOCKERFILE_TMP=".Dockerfile.tmp"
  DSEMVER=$(echo "$SEMVER" | tr '+' '-')
  DIMG_NAME="rexray/${TAG}:${DSEMVER}"
  DCNAME="rexray-$(date +%s)"
fi

if [ "$DEBUG" = "1" ]; then
  echo "SEMVER=$SEMVER"
  echo "DOCKERFILE_SRC=$DOCKERFILE_SRC"
  echo "DOCKERFILE_TMP=$DOCKERFILE_TMP"
  echo "DSEMVER=$DSEMVER"
  echo "DCNAME=$DCNAME"
  echo "DIMG_NAME=$DIMG_NAME"
fi

echo
echo "building REX-Ray (this may take a few minutes)"
echo
echo "  Builder.............. ${BUILDER}"
echo "  Version.............. ${SEMVER}"
if [ "$RR_URI" = "" ]; then
echo "  REX-Ray.............. local"
else
printf "  REX-Ray.............. %s/tree/%s\n" "$RR_URI" "$RR_REF"
fi
if [ "$LS_LOCAL" = "1" ]; then
echo "  libStorage........... local"
elif [ ! "$LS_URI" = "" ]; then
printf "  libStorage........... %s/tree/%s\n" "$LS_URI" "$LS_REF"
fi
echo

if [ "$BUILDER" = "docker" ]; then
  create_dockerfile
  if ! docker build -f "$DOCKERFILE_TMP" -t "$DIMG_NAME" .; then
    echo "error building docker image"
    exit 1
  fi
  docker create --name "$DCNAME" "$DIMG_NAME"
  docker cp "${DCNAME}:/usr/bin/${REAL_FNAME}" "$FNAME"
  rm -f "$DOCKERFILE_TMP"
  docker stop "$DCNAME" > /dev/null 2>&1 && docker rm "$DCNAME" > /dev/null 2>&1
  docker rmi $(docker images -f dangling=true -q) > /dev/null 2>&1
else
  echo "calculating make targets"
  echo "(the screen may appear frozen for a few moments)"
  echo
  if [ ! "$NOCLN" = "1" ]; then
    PORCELAIN="1" make clobber > /dev/null
  fi
  NOSTAT="1" NODOCKER="1" DRIVERS="$DRVRS" REXRAY_BUILD_TYPE="$BTYPE" make
  cp -f "${GOPATH}/bin/${REAL_FNAME}" "$FNAME"
fi

echo
echo "successfully built REX-Ray!"
echo
if [ "$BUILDER" = "docker" ]; then
echo "  Docker image is...... ${DIMG_NAME}"
fi
echo "  REX-Ray binary is.... ./${FNAME}"
echo
