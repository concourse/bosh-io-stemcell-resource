#!/usr/bin/env bats

source $BATS_TEST_DIRNAME/../assets/common.sh

load assertions

@test "it downloads a file" {
  url="https://d26ekeud912fhb.cloudfront.net/bosh-stemcell/aws/light-bosh-stemcell-3130-aws-xen-hvm-ubuntu-trusty-go_agent.tgz"
  md5="5c51e558b19f4ab59399bfafcbea7848"
  directory=$(mktemp -d $TMPDIR/bosh-io-stemcell-resource.XXXXXX)

  run download_tarball "$url" "$directory" "$md5"
  assert_success

  [ -f $directory/stemcell.tgz ]

  rm -r $directory
}

@test "it errors if the md5 sum does not match" {
  url="https://d26ekeud912fhb.cloudfront.net/bosh-stemcell/aws/light-bosh-stemcell-3130-aws-xen-hvm-ubuntu-trusty-go_agent.tgz"
  md5="this-is-not-the-correct-hash"
  directory=$(mktemp -d $TMPDIR/bosh-io-stemcell-resource.XXXXXX)

  run download_tarball "$url" "$directory" "$md5"
  assert_failure

  rm -r $directory
}

@test "it does not error if the md5 is not specified" {
  url="https://d26ekeud912fhb.cloudfront.net/bosh-stemcell/aws/light-bosh-stemcell-3130-aws-xen-hvm-ubuntu-trusty-go_agent.tgz"
  md5=""
  directory=$(mktemp -d $TMPDIR/bosh-io-stemcell-resource.XXXXXX)

  run download_tarball "$url" "$directory" "$md5"
  assert_success

  [ -f $directory/stemcell.tgz ]

  rm -r $directory
}

@test "it checks the sha1 sum if specified" {
  url="https://d26ekeud912fhb.cloudfront.net/bosh-stemcell/aws/light-bosh-stemcell-3130-aws-xen-hvm-ubuntu-trusty-go_agent.tgz"
  md5=""
  sha1="8a0958a932d9b63e813d1307d3211b17ecf451e5"
  directory=$(mktemp -d $TMPDIR/bosh-io-stemcell-resource.XXXXXX)

  run download_tarball "$url" "$directory" "$md5" "$sha1"
  assert_success

  [ -f $directory/stemcell.tgz ]

  rm -r $directory
}

@test "it errors if the sha1 sum does not match" {
  url="https://d26ekeud912fhb.cloudfront.net/bosh-stemcell/aws/light-bosh-stemcell-3130-aws-xen-hvm-ubuntu-trusty-go_agent.tgz"
  md5=""
  sha1="this-is-the-wrong-hash"
  directory=$(mktemp -d $TMPDIR/bosh-io-stemcell-resource.XXXXXX)

  run download_tarball "$url" "$directory" "$md5" "$sha1"
  assert_failure

  rm -r $directory
}

@test "uses the sha1 to check if both md5 and sha1 are specified" {
  url="https://d26ekeud912fhb.cloudfront.net/bosh-stemcell/aws/light-bosh-stemcell-3130-aws-xen-hvm-ubuntu-trusty-go_agent.tgz"
  md5="this-is-a-bad-hash"
  sha1="8a0958a932d9b63e813d1307d3211b17ecf451e5"
  directory=$(mktemp -d $TMPDIR/bosh-io-stemcell-resource.XXXXXX)

  run download_tarball "$url" "$directory" "$md5" "$sha1"
  assert_success

  rm -r $directory
}
