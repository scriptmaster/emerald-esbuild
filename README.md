# emerald-esbuild

Build client apps with esbuild and typescript without any node_modules folder.
File are pulled from https://esm.sh/<package_name@version> on the fly and built
everytime.

## Install

`go install github.com/scriptmaster/emerald-esbuild@latest`

## Usage

After go install was successful.

`emerald-esbuild`

on an empty directory to start building files.

## How it works

- `emerald-esbuild` is a wrapper on esbuild api.
- It transforms the entrypoint `app/main.tsx` and any imports within it
  (downloads via its aliases in importmap.json) and bundles into `dist/main.js`
  and `dist/main.css`
