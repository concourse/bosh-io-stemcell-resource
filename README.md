# bosh.io stemcell resource

Tracks the versions of a stemcell on [bosh.io](https://bosh.io).

<a href="https://ci.concourse-ci.org/teams/main/pipelines/resource/jobs/build?vars.type=%22bosh-io-stemcell%22">
  <img src="https://ci.concourse-ci.org/api/v1/teams/main/pipelines/resource/jobs/build/badge?vars.type=%22bosh-io-stemcell%22" alt="Build Status">
</a>

For example, to automatically consume `bosh-aws-xen-hvm-ubuntu-trusty-go_agent`:

```yaml
resources:
- name: aws-stemcell
  type: bosh-io-stemcell
  source:
    name: bosh-aws-xen-hvm-ubuntu-trusty-go_agent
```

## Source Configuration

* `name`: *Required.* The name of the stemcell.

* `version_family`: *Optional.* Default `latest`. A semantic version used to
narrow the returned versions, typically used to fetch hotfixes on older
stemcells. For example, a `version_family` of `3262.latest` would match `3262`,
`3262.1`, and `3262.1.1`, but not `3263`. A `version_family` of `3262.1.latest`
would match `3262.1` and `3262.1.1`, but not `3262.2`.

* `force_regular`: *Optional.* Default `false`. By default, the resource will always download light stemcells for IaaSes that support light stemcells.
  If `force_regular` is `true`, the resource will ignore light stemcells and always download regular stemcells.

* `force_light`: *Optional.* Default `false`. When `true`, the resource will fail during `in` if no light stemcell is available for the requested version, rather than silently falling back to the regular stemcell. Has no effect on IaaSes that do not support light stemcells. Useful when a pipeline must guarantee it never inadvertently publishes a heavy stemcell.

* `auth`: *Optional.* These credentials are used when downloading stemcells stored in a protected bucket.
  Has the following sub-properties:
  * `access_key`: *Required.* The HMAC access key
  * `secret_key`: *Required.* The HMAC secret key

* `private_bucket`: *Optional.* Discover and download stemcells from an S3-compatible private bucket instead of bosh.io metadata. This is useful for stemcells that are not published through the bosh.io API. Requires `auth`.
  Has the following sub-properties:
  * `endpoint`: *Required.* The S3-compatible service endpoint, for example `https://storage.googleapis.com`
  * `bucket`: *Required.* The bucket containing the stemcell objects
  * `regexp`: *Required.* A regular expression matched against object names. The first capture group is used as the version, or the named capture group `version` if present.

  For example:

  ```yaml
  resources:
  - name: private-stemcell
    type: bosh-io-stemcell
    source:
      name: bosh-aws-xen-hvm-ubuntu-jammy-fips-go_agent
      auth:
        access_key: ((stemcell_hmac_access_key))
        secret_key: ((stemcell_hmac_secret_key))
      private_bucket:
        endpoint: https://storage.googleapis.com
        bucket: bosh-core-stemcells-fips
        regexp: '(?P<version>[^/]+)/bosh-stemcell-[^/]+-aws-xen-hvm-ubuntu-jammy-fips-go_agent\.tgz'
  ```

## Behavior

### `check`: Check for new versions of the stemcell.

Detects new versions of the stemcell that have been published to [bosh.io](https://bosh.io). If `private_bucket` is configured, detects versions by listing matching objects from that bucket. If no version is specified, `check` returns the latest version, otherwise `check` returns all versions from the version specified on.


### `in`: Fetch a version of the stemcell.

Fetches a given stemcell, placing the following files in the destination:

* `version`: The version number of the stemcell.
* `url`: A URL that can be used to download the stemcell tarball.
* `sha1`: The SHA1 of the stemcell
* `sha256`: The SHA256 of the stemcell
* `stemcell.tgz`: The stemcell tarball, if the `tarball` param is `true`.

When `private_bucket` is configured, `sha1` and `sha256` are computed after downloading the tarball. If `tarball` is `false`, checksum files may be empty because bosh.io metadata is not used.

#### Parameters

* `tarball`: *Optional.* Default `true`. Fetch the stemcell tarball.
* `preserve_filename`: *Optional.* Default `false`. Keep the original filename of the stemcell.

## Development

### Prerequisites

* golang is *required* - version 1.9.x is tested; earlier versions may also
  work.
* docker is *required* - version 23.x is tested; earlier versions may also
  work.

### Running the tests

The tests have been embedded with the `Dockerfile`; ensuring that the testing
environment is consistent across any `docker` enabled platform. When the docker
image builds, the test are run inside the docker container, on failure they
will stop the build.

Run the tests with the following command:

```sh
docker build -t bosh-io-stemcell-resource --target tests .
```

### Contributing

Please make all pull requests to the `master` branch and ensure tests pass
locally.
