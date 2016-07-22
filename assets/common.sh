download_tarball() {
  local url=$1
  local destination=$2
  local md5=$3
  local sha1=$4

  curl --retry 5 -s -f "$url" -o $destination

  if [ -z "$md5" -a -z "$sha1" ]; then
    return 0
  fi

  if [ ! -z "$sha1" ]; then
    fetched_sha=$(sha1_of $destination)
    if [ "$fetched_sha" != "$sha1" ]; then
      echo "checksum mismatch: want $sha1, got $fetched_sha"
      return 1
    fi

    return 0
  fi

  if [ ! -z "$md5" ]; then
    fetched_sum=$(md5_of $destination)
    if [ "$fetched_sum" != "$md5" ]; then
     echo "checksum mismatch: want $md5, got $fetched_sum"
     return 1
    fi
  fi
}

sha1_of() {
  local file=$1

  if command_exists sha1sum; then
    sha1sum $file | awk '{print $1}'
  elif command_exists shasum; then
    shasum $file | awk '{print $1}'
  else
    echo "no sha1 checksum program installed!"
    exit 1
  fi
}

md5_of() {
  local file=$1

  if command_exists md5sum; then
    md5sum $file | awk '{print $1}'
  elif command_exists md5; then
    md5 $file | awk '{print $4}'
  else
    echo "no md5 checksum program installed!"
    exit 1
  fi
}

command_exists() {
  command -v $1 >/dev/null 2>&1
}
