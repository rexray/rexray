#!/bin/sh

path="${1:-.}"
size="${2:-1}"
unit="${3:-M}" # ckMGTP
if [ "$unit" = "b" ]; then unit="c"; fi

find_big_files() {
  find "$path" \
    -not -path '*/.docker/*' \
    -not -path '*/.docs/*' \
    -not -path '*/.git/*' \
    -not -path '*/.site/*' \
    -not -path '*/vendor/*' \
    -type f \
    -size +"${size}${unit}" 2> /dev/null
}

print_big_files() {
  if [ "$#" -eq 0 ]; then return 1; fi
  $(env which ls) -lhS $* | \
  awk '{ sub("'"$path"'\/?","",$9); print $5 "\t" $9 }' 2> /dev/null
}

print_big_files $(find_big_files)
