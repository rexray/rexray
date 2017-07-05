#!/bin/sh

################################################################################
# build REX-Ray
#

ekho() {
    echo "$@" 1>&2
}
usage() {
  ekho "usage: build.sh [-l loglevel] [-b builder] [-x] [-t type] [-d drivers]"
  ekho "                [-a arch] [-e executors]"
  ekho "                [-u rr_uri] [-r rr_ref]"
  ekho "                [-s] | [-1 ls_uri] [-2 ls_ref]"
  ekho "                [-3] [FILE]"
  ekho
  ekho "       -l loglevel      Optional. Sets the program's log level. Possible"
  ekho "                        values include debug, info, warn, and error. The"
  ekho "                        default log level is error."
  ekho
  ekho "       -a arch          Optional. Sets the GOARCH environment variable."
  ekho "                        Defaults to amd64."
  ekho
  ekho "       -e exeuctors     Optional. A quoted, space delimited list of the"
  ekho "                        GOOS_GOARCH executors to embed. The default value"
  ekho "                        is GOOS_GOARCH."
  ekho
  ekho "       -b builder       Optional. The builder used to build REX-Ray. Possible"
  ekho "                        values include docker, make & skip. The default value,"
  ekho "                        if installed and running, is docker. If Docker is not"
  ekho "                        installed or the current user cannot access it then"
  ekho "                        make is used."
  ekho
  ekho "                        If docker is *explicitly* specified but not installed"
  ekho "                        or inaccessible, this program will *not* switch to"
  ekho "                        make, but instead exit with an error."
  ekho
  ekho "                        Using skip to skip the build is useful for when"
  ekho "                        wanting to build a plug-in without rebuilding"
  ekho "                        its artifact."
  ekho
  ekho "       -x               Optional. This flag is only applicable to the make"
  ekho "                        build runner. If set, this flag prevents make from"
  ekho "                        performing a clean ahead of the build. Therefore"
  ekho "                        this flag results in the preservation of e-X-isting"
  ekho "                        files."
  ekho
  ekho "                        This flag sets '-b make'."
  ekho
  ekho "       -t type          Optional. The type of REX-Ray binary to create."
  ekho "                        Possible values include: agent, client, controller, and"
  ekho "                        plugin."
  ekho
  ekho "                        The default value is the same as omitting this flag and"
  ekho "                        argument altogether and will create a binary that"
  ekho "                        includes the agent, client, and controller."
  ekho
  ekho "       -d drivers       Optional. One or more drivers to include in the binary."
  ekho "                        Specify multiple drivers as a quoted, space-delimited"
  ekho "                        list."
  ekho
  ekho "                        Drivers are only included in standard and controller"
  ekho "                        binaries. Please see the -t type flag for more"
  ekho "                        information on binary types."
  ekho
  ekho "       -u uri           Optional. The URI to a REX-Ray, git repository . Thisfffff"
  ekho "                        flag is only valid for the docker builder. Specifying"
  ekho "                        a repository will use its sources (specified by the"
  ekho "                        associated -r ref flag) instead of the local sources."
  ekho
  ekho "       -r ref           Optional. The git references to use when specifying the"
  ekho "                        -u uri flag. This value can be a commit ID, tag, or"
  ekho "                        branch name. The default value is master."
  ekho
  ekho "                        This flag sets '-b docker'."
  ekho
  ekho "                        If -r is set but -u is not, -u will be set to"
  ekho "                        to libStorage's primary repository URI."
  ekho
  ekho "       -s               Optional. A flag indicating to use the local libStorage"
  ekho "                        sources instead of the libStorage version specified in"
  ekho "                        REX-Ray's glide.yaml file."
  ekho
  ekho "                        This flag sets '-b docker' and cannot be used with the"
  ekho "                        -1 or -2 flags."
  ekho
  ekho "       -1 uri           Optional. The URI to a libStorage, git repository. This"
  ekho "                        flag is only valid for the docker builder. Specifying"
  ekho "                        a repository will use its sources (specified by the"
  ekho "                        associated -2 ref flag) instead of the libStorage"
  ekho "                        version specified in REX-Ray's glide.yaml file."
  ekho
  ekho "                        This flag sets '-b docker' and cannot be used with the"
  ekho "                        -s flag."
  ekho
  ekho "       -2 ref           Optional. The git references to use when specifying the"
  ekho "                        -1 uri flag. This value can be a commit ID, tag, or"
  ekho "                        branch name. The default value is master."
  ekho
  ekho "                        This flag sets '-b docker' and cannot be used with the"
  ekho "                        -s flag."
  ekho
  ekho "                        If -2 is set but -1 is not, -1 will be set to"
  ekho "                        to libStorage's primary repository URI."
  ekho
  ekho "       -3               Optional. This is a flag that indicates not to keep "
  ekho "                        the Docker image that is the result of a Docker build."
  ekho
  ekho "                        This flag is ignored when -b make."
  ekho
  ekho "       FILE             Optional. The name of the produced binary. The default"
  ekho "                        name of the binary is based on the binary type. For"
  ekho "                        example, if -t client is set then the file will be"
  ekho "                        rexray-client. If no type is set then the default file"
  ekho "                        name is rexray."
  exit 1
}

# the version of go to use
GO_VERSION="${TRAVIS_GO_VERSION:-$(grep -A 1 '^go:' .travis.yml | \
  tail -n 1 | \
  awk '{print $2}')}"

# the makeflags to use for make commands
MAKEFLAGS=--no-print-directory

# a timestamp constant for the program
EPOCH="$(date +%s)"

# check to see if docker is present and running
if docker version > /dev/null 2>&1; then
  DOCKER="1"
fi

# a flag indicating the log level. the log levels are:
#
#   debug  1
#   info   2
#   warn   3
#   error  4
#
DEBUG="${DEBUG:-4}"
parse_loglevel() {
  if echo "$DEBUG" | grep -i '^\(1\|true\|debug\)$' > /dev/null 2>&1; then
    DEBUG="1"
    KEEP_TEMP_FILES="${KEEP_TEMP_FILES:-1}"
  elif echo "$DEBUG" | grep -i  '^\(2\|info\)$' > /dev/null 2>&1; then
    DEBUG="2"
  elif echo "$DEBUG" | grep -i  '^\(3\|warn\)$' > /dev/null 2>&1; then
    DEBUG="3"
  else
    DEBUG="4"
  fi
}

# log adheres to sprintf
log() {

  # if no log level was specified then get out
  if [ "$#" -eq 0 ]; then return 0; fi

  level="$1"
  shift

  # do not log if the entry's log level is less
  # than the program's log level
  if [ "$level" -lt "$DEBUG" ]; then return 0; fi

  case $level in
  1)
    szlevel="DEBUG"
    ;;
  2)
    szlevel="INFO"
    ;;
  3)
    szlevel="WARN"
    ;;
  4)
    szlevel="ERROR"
    ;;
  esac

  printf "[%s]\t" "$szlevel" 1>&2

  # if there are no more arguments then print
  # a blank line and get out
  if [ "$#" -eq 0 ]; then printf "\n" 1>&2 && return 0; fi

  # if there is only one argument then print it
  # and get out
  if [ "$#" -eq 1 ]; then printf "%s\n" "$1" 1>&2 && return 0; fi

  # there are multiple arguments. treat the first
  # argument as the format string and then shift
  # the arguments by one, using the remaining args
  # as the format string's args with printf
  format="$1"
  shift
  # shellcheck disable=SC2059
  printf "${format}\n" "$@" 1>&2
}

# log when DEBUG=1|true
debug() {
  log "1" "$@"
}

# log when DEBUG=2
info() {
  log "2" "$@"
}

# log when DEBUG=3
warn() {
  log "3" "$@"
}

# log when DEBUG=4
error() {
  log "4" "$@"
}

cleanup() {
  if [ "$DOCKER" = "1" ] && [ "$BUILDER" = "docker" ]; then
    debug "cleaning up docker build"
    if [ "$DCNAME" != "" ]; then
      debug "stopping docker container: %s" "$DCNAME"
      docker stop "$DCNAME" > /dev/null 2>&1
      debug "removing docker container: %s" "$DCNAME"
      docker rm "$DCNAME" > /dev/null 2>&1
    fi
    if [ "$NOKEEP" = "1" ] && [ "$DIMG_NAME" != "" ]; then
      debug "removing docker image: %s" "$DIMG_NAME"
      docker rmi "$DIMG_NAME" > /dev/null 2>&1
    fi
  fi
  if [ "$DOCKER" = "1" ] && \
     ([ "$BUILDER" = "docker" ] || [ "$BTYPE" = "plugin" ]); then
    debug "removing dangling docker images"
    docker rmi "$(docker images -f dangling=true -q)" > /dev/null 2>&1
  fi
  if [ "$KEEP_TEMP_FILES" != "1" ]; then
    if [ -f "$DOCKERFILE_TMP" ]; then
      debug "removing temp dockerfile: %s" "$DOCKERFILE_TMP"
      rm -f "$DOCKERFILE_TMP"
    fi
    if [ -f "$GLIDE_YAML_TMP" ]; then
      debug "removing temp glide config: %s" "$GLIDE_YAML_TMP"
      rm -f "$GLIDE_YAML_TMP"
    fi
    if [ -f "$LIBSTORAGE_TGZ" ]; then
      debug "removing temp libstorage tarball: %s" "$LIBSTORAGE_TGZ"
      rm -f "$LIBSTORAGE_TGZ"
    fi
  fi
}

# ensure the cleanup function is always executed upon exit
trap cleanup EXIT

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
    if ! version="$(PORCELAIN=1 make version-porcelain)"; then return 1; fi
    echo "$version" | tr -d "\r" | tr -d "\n"
    return 0
  fi

  rr_uri="$(get_git_repo "$2")"
  rr_ref="${3:-master}"
  go_ver="$GO_VERSION"

  vcmd="PORCELAIN=1 make $MAKEFLAGS -C /rexray version-porcelain"
  if [ "$rr_uri" = "" ]; then
    if ! version=$(docker run -it --rm -v "$(pwd)":/rexray \
         golang:${go_ver} bash -c ''"$vcmd"''); then
         return 1
    fi
    echo "$version" | tr -d "\r" | tr -d "\n"
    return 0
  fi

  mkgd=" mkdir -p /rexray && cd /rexray"
  vcmd="$mkgd && $(git_clone_checkout_cmds "$rr_uri" "$rr_ref") && $vcmd"
  if ! version=$(docker run -it --rm golang:${go_ver} bash -c ''"$vcmd"''); then
    return 1
  fi
  echo "$version" | tr -d "\r" | tr -d "\n"
  return 0
}

create_dockerfile() {
  workdir_rr="/go/src/github.com/codedellemc/rexray/"
  bsrc='GOARCH="'"${GOARCH}"'" NOSTAT="1" DRIVERS="'"${DRIVERS}"'" '
  bsrc="$bsrc"'EMED_EXECUTORS="'"${EMED_EXECUTORS}"'" '
  bsrc="$bsrc"'REXRAY_BUILD_TYPE='"${BTYPE}"' make'

  lsd="${GOPATH}/src/github.com/codedellemc/libstorage"
  if [ "$LS_LOCAL" = "1" ] && [ -d "$lsd" ]; then
    tar -czf "$LIBSTORAGE_TGZ" \
      --exclude "./.git" \
      --exclude "./vendor" \
      --exclude "*.test" \
      -C "$lsd" .

    ls_home="/go/src/github.com/codedellemc/libstorage/"
    workdir_ls="WORKDIR $ls_home"

    nl="$(printf '%b_' '\n')"
    sed -e 's|.*# libstorage-version|    ref:     master|g' \
        -e 's|.*# libstorage-repo|    repo:    file://'"${ls_home}\\${nl%_}"'    vcs:     git|g' \
        glide.yaml > "$GLIDE_YAML_TMP"
    bsrc="cp -f $GLIDE_YAML_TMP glide.yaml \&\& $bsrc"

    init_ls_srcs="RUN tar xzf ${workdir_rr}${LIBSTORAGE_TGZ} \&\& "
    init_ls_srcs="${init_ls_srcs}git init \&\& "
    init_ls_srcs="${init_ls_srcs}git config --local user.name $(whoami) \&\& "
    init_ls_srcs="${init_ls_srcs}git config --local user.email $(whoami)@localhost \&\& "
    init_ls_srcs="${init_ls_srcs}git add -A > /dev/null \&\& "
    init_ls_srcs="${init_ls_srcs}git commit -m v0.0.1 > /dev/null \&\& "
    init_ls_srcs="${init_ls_srcs}git tag -a -m v0.0.1 v0.0.1"

  elif [ "$LS_URI" != "" ]; then
     nl="$(printf '%b_' '\n')"
     sed -e 's/.*# libstorage-version/    ref:     '"$LS_REF"'/g' \
         -e 's|.*# libstorage-repo|    repo:    '"$LS_URI\\${nl%_}"'    vcs:     git|g' \
         glide.yaml > "$GLIDE_YAML_TMP"
     bsrc="cp -f $GLIDE_YAML_TMP glide.yaml \&\& $bsrc"
  fi

  if [ "$RR_URI" = "" ]; then
    copy_rr_srcs='COPY . .'
    bcmd="$bsrc"
  else
    if [ "$LS_LOCAL" = "1" ]; then
      copy_rr_srcs='COPY ['"$LIBSTORAGE_TGZ"', '"$GLIDE_YAML_TMP"', "./"]'
    elif [ "$LS_URI" != "" ]; then
      copy_rr_srcs='COPY '"$GLIDE_YAML_TMP"' .'
    fi
    bgit='(mv '"$GLIDE_YAML_TMP"' /tmp/ > /dev/null 2>\&1 || true)'
    bgit="$bgit"' \&\& git clone '"$RR_URI"' . \&\& git checkout '"$RR_REF"
    bgit="$bgit"' \&\& if git status | grep "HEAD detached" > /dev/null; '
    bgit="$bgit"'then git checkout -b '"$RR_REF"'; fi'
    bgit="$bgit"' \&\& (mv /tmp/'"$GLIDE_YAML_TMP"' . > /dev/null 2>\&1 || true)'
    bcmd="$bgit"' \&\& '"$bsrc"
  fi

  if [ "$GOOS" != "" ] || [ "$GOARCH" != "" ]; then
    GOOS="${GOOS:-linux}"
    GOARCH="${GOARCH:-amd64}"
    if [ "$GOOS" != "linux" ] || [ "$GOARCH" != "amd64" ]; then
      GOOS_GOARCH_DIR="${GOOS}_${GOARCH}/"
    fi
  fi

  sed -e 's/@GO_VERSION@/'"$GO_VERSION"'/g' \
    -e 's|@GOOS_GOARCH_DIR@|'"$GOOS_GOARCH_DIR"'/|g' \
    -e 's|@WORKDIR_RR@|'"$workdir_rr"'|g' \
    -e 's|@WORKDIR_LS@|'"$workdir_ls"'|g' \
    -e 's|@INIT_LS_SRCS_CMD@|'"$init_ls_srcs"'|g' \
    -e 's|@COPY_RR_SRCS_CMD@|'"$copy_rr_srcs"'|g' \
    -e 's%@BUILD_CMD@%'"$bcmd"'%g' \
    -e 's/@BUILD_TYPE@/'"$BTYPE"'/g' \
    -e 's/@FNAME_SUFFIX@/'"$FNAME_SUFFIX"'/g' \
    -e 's|@SEMVER@|'"$SEMVER"'|g' \
    -e 's|@DRIVERS@|'"$DRIVERS"'|g' \
    -e 's/@DOCKERFILE@/'"$DOCKERFILE_TMP"'/g' \
    "$DOCKERFILE_SRC" > "$DOCKERFILE_TMP"

  return 0
}

build_docker() {
  if ! create_dockerfile; then
    return 1
  fi
  if ! docker build -f "$DOCKERFILE_TMP" -t "$DIMG_NAME" .; then
    ekho "error building docker image $DIMG_NAME from $DOCKERFILE_TMP"
    return 1
  fi
  if ! docker create --name "$DCNAME" "$DIMG_NAME"; then
    ekho "error create container $DCNAME from $DIMG_NAME"
    return 1
  fi
  if ! docker cp "${DCNAME}:/usr/bin/${REAL_FNAME}" "$FNAME"; then
    printf "error copying build artfact %s from %s to %s\n" \
           "/usr/bin/${REAL_FNAME}", "$DCNAME", "$FNAME"
    return 1
  fi
  return 0
}

build_make() {
  echo "calculating make targets"
  echo "(the screen may appear frozen for a few moments)"
  echo
  if [ "$NOCLEAN" != "1" ]; then
    if ! PORCELAIN="1" make clobber > /dev/null; then
      ekho "error cleaning up prior to make"
      return 1
    fi
  fi
  if ! GOARCH="$GOARCH" NOSTAT="1" NODOCKER="1" \
       EMBED_EXECUTORS="$EMBED_EXECUTORS" \
       DRIVERS="$DRIVERS" REXRAY_BUILD_TYPE="$BTYPE" make; then
    ekho "error building with make"
    return 1
  fi
  if [ ! -f "$REAL_FPATH" ]; then
    ekho "error: build artifact missing: $REAL_FPATH"
    return 1
  fi
  cp -f "$REAL_FPATH" "$FNAME"
  return 0
}

# ensure_plugin_files creates a plug-in's README.md and config.json if
# they do not exist in the directory .docker/plugins/${DRIVERS}
ensure_plugin_files() {
  pdir=".docker/plugins/${FDRIVERS}"
  preadme="${pdir}/README.md"
  pconfig="${pdir}/config.json"

  mkdir -p "$pdir"
  if [ ! -f "$preadme" ]; then
cat << EOF > "$preadme"
# REX-Ray Docker Plug-in for ${DRIVERS}
EOF
  fi

  if [ ! -f "$pconfig" ]; then
cat << EOF > "$pconfig"
{
      "Args": {
        "Description": "",
        "Name": "",
        "Settable": null,
        "Value": null
      },
      "Description": "REX-Ray for ${DRIVERS}",
      "Documentation": "https://github.com/codedellemc/rexray/${pdir}",
      "Entrypoint": [
        "/rexray.sh", "rexray", "start", "-f", "--nopid"
      ],
      "Env": [
        {
          "Description": "",
          "Name": "REXRAY_FSTYPE",
          "Settable": [
            "value"
          ],
          "Value": "ext4"
        },
        {
          "Description": "",
          "Name": "REXRAY_LOGLEVEL",
          "Settable": [
            "value"
          ],
          "Value": "warn"
        },
        {
          "Description": "",
          "Name": "REXRAY_PREEMPT",
          "Settable": [
            "value"
          ],
          "Value": "false"
        }
      ],
      "Interface": {
        "Socket": "rexray.sock",
        "Types": [
          "docker.volumedriver/1.0"
        ]
      },
      "Linux": {
        "AllowAllDevices": true,
        "Capabilities": ["CAP_SYS_ADMIN"],
        "Devices": null
      },
      "Mounts": [
        {
          "Source": "/dev",
          "Destination": "/dev",
          "Type": "bind",
          "Options": ["rbind"]
        }
      ],
      "Network": {
        "Type": "host"
      },
      "PropagatedMount": "/var/lib/libstorage/volumes",
      "User": {},
      "WorkDir": ""
}
EOF
  fi
}

# build_plugin builds a managed docker plug-in with the build artifact
build_plugin() {
  if ! ensure_plugin_files; then
    ekho "error validating/creating plug-in files"
    return 1
  fi

  echo "calculating make targets"
  echo "(the screen may appear frozen for a few moments)"
  echo

  # build the plug-in using the rexray binary that was just built
  DOCKER_PLUGIN_REXRAYFILE="$FNAME" DRIVERS="$DRIVERS" make docker-build-plugin
}

################################################################################
##                                                                            ##
##                                main(argv, argc)                            ##
##                                                                            ##
################################################################################

# the builder
BUILDER="docker"

# the go architecture to build
GOARCH="${GOARCH:-amd64}"

# the build type
BTYPE="${REXRAY_BUILD_TYPE:-}"

# the drivers
DRIVERS="${DRIVERS:-}"

# the file name
FNAME=

# do not clean ahead of a make
NOCLEAN=

# do not keep the image used to build REX-Ray
NOKEEP=

while getopts ":l:a:e:b:t:d:xu:r:s1:2:3" opt; do
  case $opt in
  l)
    FLAG_L="1"
    DEBUG="$OPTARG"
    ;;
  a)
    FLAG_A="1"
    GOARCH="$OPTARG"
    ;;
  e)
    FLAG_E="1"
    EMED_EXECUTORS="$OPTARG"
    ;;
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
    DRIVERS="$OPTARG"
    ;;
  x)
    FLAG_X="1"
    NOCLEAN="1"
    ;;
  u)
    FLAG_U="1"
    RR_URI="$OPTARG"
    ;;
  r)
    FLAG_R="1"
    RR_REF="$OPTARG"
    ;;
  s)
    FLAG_S="1"
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
  3)
    FLAG_3="1"
    NOKEEP="1"
    ;;
  *)
    usage
    ;;
  esac
done
shift $((OPTIND-1))

# if -l was set parse the log level
if [ "$FLAG_L" = "1" ]; then parse_loglevel; fi

# validate the builder
if ! echo "$BUILDER" | grep -i '^\(docker\|make\|skip\)$' > /dev/null 2>&1; then
  ekho "error: invalid builder: $BUILDER"
  ekho
  usage
fi
BUILDER="$(echo "$BUILDER" | tr '[:upper:]' '[:lower:]')"

# if the docker builder was selected then make sure it is available
if [ "$BUILDER" = "docker" ]; then
  if [ "$DOCKER" != "1" ]; then
    if [ "$FLAG_B" = "1" ]; then
      ekho "error: docker builder unavailable"
      ekho
      usage
    fi
    BUILDER="make"
  fi
fi

# validate the libstorage source location
if [ "$FLAG_S" = "1" ] && \
   ([ "$FLAG_1" = "1" ] || [ "$FLAG_2" = "1" ]); then
  ekho "error: cannot use both local & remote libStorage sources"
  ekho
  usage
fi

# validate that no make-only flags were specified with a docker builder
if [ "$BUILDER" = "docker" ] && [ "$FLAG_X" = "1" ]; then
  ekho "error: -x cannot be used with the docker builder"
  ekho
  usage
fi

# validate that no docker-only flags were specified with a make builder
if [ "$BUILDER" = "make" ] && \
   ([ "$FLAG_U" = "1" ] || [ "$FLAG_R" = "1" ] || \
    [ "$FLAG_S" = "1" ] || [ "$FLAG_1" = "1" ] || \
    [ "$FLAG_2" = "1" ]); then
  ekho "error: -u,-r,-s,-1,-2 cannot be used with the make builder"
  ekho
  usage
fi

# validate the build type
if [ "$BTYPE" != "" ] && \
   [ "$BTYPE" != "agent" ] && \
   [ "$BTYPE" != "client" ] && \
   [ "$BTYPE" != "controller" ] && \
   [ "$BTYPE" != "plugin" ]; then
   ekho "error: invalid build type: $BTYPE"
   ekho
   usage
fi

# if the -u flag was set then sanitize the uri
if [ "$FLAG_U" = "1" ]; then
  RR_URI="$(get_git_repo "$RR_URI" rexray)"
fi

# if the -1 flag was set then sanitize the uri
if [ "$FLAG_1" = "1" ]; then
  LS_URI="$(get_git_repo "$LS_URI" libstorage)"
fi

# if there is a rex-ray uri set, ensure the ref is defined
if [ "$RR_URI" != "" ]; then
  RR_REF="${RR_REF:-master}"
# if there is a rex-ray ref set, ensure the uri is defined
elif [ "$RR_REF" != "" ]; then
  RR_URI="https://github.com/codedellemc/rexray"
fi

# if there is a libstorage uri set, ensure the ref is defined
if [ "$LS_URI" != "" ]; then
  LS_REF="${LS_REF:-master}"
# if there is a libstorage ref set, ensure the uri is defined
elif [ "$LS_REF" != "" ]; then
  LS_URI="https://github.com/codedellemc/libstorage"
fi

# validate that drivers aren't set for agent or client builds
if [ "$DRIVERS" != "" ] && \
  ([ "$BTYPE" = "agent" ] || [ "$BTYPE" = "client" ]); then
  ekho "error: drivers are invalid for agent & client builds"
  ekho
  usage
# require one or more drivers for a plugin build
elif [ "$DRIVERS" = "" ] && [ "$BTYPE" = "plugin" ]; then
  ekho "error: must specify a driver when building a plug-in"
  ekho
  usage
fi

# when building a plug-in explicitly set the build type to nothing
# and indicate that a plug-in was requested
if [ "$BTYPE" = "plugin" ]; then
  if [ "$DOCKER" != "1" ]; then
    ekho "error: the artifact used to create the plug-in can be built with "
    ekho "       make, but building the plug-in requires Docker"
    exit 1
  fi

  FDRIVERS=$(echo "$DRIVERS" | tr ' ' '-' )
  BTYPE=""
  BPLUG="1"
  FNAME="rexray-${FDRIVERS}"
  REAL_FNAME="rexray"

  # a plug-in build automatically removes the docker image
  # used to build the REX-Ray binary
  NOKEEP="1"
elif [ "$BTYPE" = "" ]; then
  FNAME="rexray"
  REAL_FNAME="rexray"
else
  FNAME="rexray-${BTYPE}"
  REAL_FNAME="rexray-${BTYPE}"
  FNAME_SUFFIX="-${BTYPE}"
fi

if [ "$GOARCH" = "amd64" ]; then
  REAL_FPATH="${GOPATH}/bin/${REAL_FNAME}"
else
  REAL_FPATH="${GOPATH}/bin/linux_${GOARCH}/${REAL_FNAME}"

  # remove the Docker image since it's incompatible with
  # the REX-Ray binary
  NOKEEP="1"
fi

# indicate which executors to embed
EMED_EXECUTORS="${EMED_EXECUTORS:-linux_$GOARCH}"

if [ "$1" != "" ]; then
  FNAME="$1"
fi

debug "FLAG_A=$FLAG_A"
debug "FLAG_B=$FLAG_B"
debug "FLAG_E=$FLAG_E"
debug "FLAG_T=$FLAG_T"
debug "FLAG_D=$FLAG_D"
debug "FLAG_X=$FLAG_X"
debug "FLAG_U=$FLAG_U"
debug "FLAG_R=$FLAG_R"
debug "FLAG_S=$FLAG_S"
debug "FLAG_1=$FLAG_1"
debug "FLAG_2=$FLAG_2"
debug "FLAG_3=$FLAG_3"
debug "GOARCH=$GOARCH"
debug "EMED_EXECUTORS=$EMED_EXECUTORS"
debug "NOCLEAN=$NOCLEAN"
debug "NOKEEP=$NOKEEP"
debug "BUILDER=$BUILDER"
debug "RR_URI=$RR_URI"
debug "RR_REF=$RR_REF"
debug "LS_URI=$LS_URI"
debug "LS_REF=$LS_REF"
debug "LS_LOCAL=$LS_LOCAL"
debug "BTYPE=$BTYPE"
debug "BPLUG=$BPLUG"
debug "DRIVERS=$DRIVERS"
debug "FNAME=$FNAME"
debug "REAL_FNAME=$REAL_FNAME"
debug "REAL_FPATH=$REAL_FPATH"
debug "FNAME_SUFFIX=$FNAME_SUFFIX"

SEMVER="${SEMVER:-$(get_semver "$BUILDER" "$RR_URI" "$RR_REF")}"

if [ "$BUILDER" = "docker" ]; then
  DOCKERFILE_SRC=".Dockerfile"
  DOCKERFILE_TMP=".Dockerfile.tmp"
  GLIDE_YAML_TMP=".glide.yaml.tmp"
  LIBSTORAGE_TGZ=".ls.tar.gz"
  DSEMVER=$(echo "$SEMVER" | tr '+' '-')
  if [ "$BTYPE" != "" ]; then
    DIMG_NAME="rexray/${BTYPE}:${DSEMVER}"
  else
    DIMG_NAME="rexray/rexray:${DSEMVER}"
  fi
  DCNAME="rexray-${EPOCH}"
fi

debug "SEMVER=$SEMVER"
debug "DOCKERFILE_SRC=$DOCKERFILE_SRC"
debug "DOCKERFILE_TMP=$DOCKERFILE_TMP"
debug "DSEMVER=$DSEMVER"
debug "DCNAME=$DCNAME"
debug "DIMG_NAME=$DIMG_NAME"

echo
echo "building REX-Ray (this may take a few minutes)"
echo
echo "  Builder.............. ${BUILDER}"
echo "  Version.............. ${SEMVER}"
if [ "$BPLUG" = "1" ]; then
echo "  Plug-in.............. Yes"
fi
if [ "$RR_URI" = "" ]; then
echo "  REX-Ray.............. local"
else
printf "  REX-Ray.............. %s/tree/%s\n" "$RR_URI" "$RR_REF"
fi
if [ "$LS_LOCAL" = "1" ]; then
echo "  libStorage........... local"
elif [ "$LS_URI" != "" ]; then
printf "  libStorage........... %s/tree/%s\n" "$LS_URI" "$LS_REF"
fi
echo

case $BUILDER in
docker)
  if ! build_docker; then exit 1; fi
  ;;
make)
  if ! build_make; then exit 1; fi
  ;;
skip)
  # do nothing
  ;;
*)
  usage
  ;;
esac

# if building a plug-in, invoke build_plugin
if [ "$BPLUG" = "1" ]; then
  if ! build_plugin; then exit 1; fi
fi

echo
echo "successfully built REX-Ray!"
echo
if [ "$BUILDER" = "docker" ] && [ "$NOKEEP" = "" ]; then
echo "  Docker image is...... ${DIMG_NAME}"
fi
echo "  REX-Ray binary is.... ./${FNAME}"
echo

exit 0
