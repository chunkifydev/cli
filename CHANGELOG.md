# Changelog

## [1.0.0](https://github.com/chunkifydev/cli/compare/v0.9.0...v1.0.0) (2026-11-17)


### ⚠ BREAKING CHANGES

* the --json progress lines no longer include the fps and out_time keys. They were always 0 because the Transcodings API doesn't report encoder statistics. Each line now has status, progress, speed and eta.
* the hls_h264, hls_h265, hls_av1 and jpg formats, the --transcoders, --vcpu, --hls-*, --interval and --sprite flags, and the HLS/VTT post-processing are removed because the Transcodings API does not support them. Generated external output paths move from chunkify-cli/jobs to chunkify-cli/transcodings.

### Features

* remove fps and out_time from JSON output ([cbf02e7](https://github.com/chunkifydev/cli/commit/cbf02e7e390a4d963b7e7c57bcee75c3c273fe4f))
* remove HLS, JPG and transcoder options ([2aab837](https://github.com/chunkifydev/cli/commit/2aab83710757628dbce633ee0d87a9e9720bc752))
* use chunkify-go v1.0.0 ([249ec23](https://github.com/chunkifydev/cli/commit/249ec23c3ad6de2204dd507ac50ff999cd6e30d3))
* use the Transcodings API and chunkify-go v0.17.0 ([d63aa3b](https://github.com/chunkifydev/cli/commit/d63aa3b68fdc94c58595a7265174951a55bc9c66))


### Bug Fixes

* make --crf optional for webm_vp9 ([5211773](https://github.com/chunkifydev/cli/commit/521177373446df5ffe50d2ad9c0301cccac253f4))
* make the transcode command work in scripts and CI ([7371019](https://github.com/chunkifydev/cli/commit/737101930ea02dddd52551222a09544256b28598))
* make the transcode command work in scripts and CI ([0c48d3d](https://github.com/chunkifydev/cli/commit/0c48d3d826d79027b08db364d6d4f22cd125a500))
* report JSON errors consistently and run without a keyring ([70d8a77](https://github.com/chunkifydev/cli/commit/70d8a77be8a46dbc05022563583f8f6745d6da32))
* show errors reliably, hide signed URLs and check output extensions ([f2d7604](https://github.com/chunkifydev/cli/commit/f2d7604a9b5a6879e41fb351d252d196fe74746e))
* stop polling on permanent errors and align flag help ([efac7ec](https://github.com/chunkifydev/cli/commit/efac7ecfa96de6934a62e0d598724a3f19769ceb))

## [0.9.0](https://github.com/chunkifydev/cli/compare/v0.8.0...v0.9.0) (2026-09-24)


### Features

* add OpenAPI-driven API commands ([#7](https://github.com/chunkifydev/cli/issues/7)) ([2603101](https://github.com/chunkifydev/cli/commit/2603101a9366d60001b5b4e56a1b58ffa541a8f2))

## [0.8.0](https://github.com/chunkifydev/cli/compare/v0.7.0...v0.8.0) (2026-09-16)


### Features

* support npm and npx installation ([#5](https://github.com/chunkifydev/cli/issues/5)) ([ee36f3e](https://github.com/chunkifydev/cli/commit/ee36f3e18d2f0871886b98e1c2f664997836a94f))

## [0.7.0](https://github.com/chunkifydev/cli/compare/v0.6.0...v0.7.0) (2026-09-16)


### Features

* support storage workflows and upload completion ([#3](https://github.com/chunkifydev/cli/issues/3)) ([09ed4f2](https://github.com/chunkifydev/cli/commit/09ed4f2f6452653e6387afb797e3389f02f7cbe3))
