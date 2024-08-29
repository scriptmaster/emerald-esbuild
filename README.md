# emerald-esbuild

Build client apps with esbuild and typescript without any node_modules folder.
Files are downloded from https://esm.sh/<package_name@version> on the fly and
built everytime.

## Install

`go install github.com/scriptmaster/emerald-esbuild@latest`

This installs to `$GOPATH/bin` dir. Ensure GOPATH was set to `$HOME/go` and
`$GOPATH/bin` was added to your `PATH` env var.

## Pre-requisites

- Install go: `https://go.dev/doc/install`
- Usually, `GOROOT` is `/usr/local/go` and `GOPATH` is `$HOME/go` for user
  packages. Do, `go env` to list GOROOT and GOPATH.

## Usage

After go install was successful.

`emerald-esbuild`

on an empty directory to start building files.

## How it works

- `emerald-esbuild` is a wrapper on esbuild api.
- It transforms the entrypoint `app/main.tsx` and any imports within it
  (downloads via its aliases in importmap.json) and bundles into `dist/main.js`
  and `dist/main.css`
