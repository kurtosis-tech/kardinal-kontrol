# kontrol-frontend

## Quickstart

```bash
nix develop
bun dev
```

The frontend can be accessed at `http://localhost:5173/<tenant uuid>`.

[Storybook](https://storybook.js.org/) can also be started using:

```bash
bun run storybook
```

## Environment variables
The frontend uses the `VITE_API_URL` in `.env.dev` to communicate with the
kontrol service. Alternatively, there is a localhost API proxy that will proxy requests to the
production API if desired. To use this, uncomment the commented line in
`.env.dev`

## Running the API locally

The kontrol service needs to be started in dev mode with the
flag `-dev-mode` so CORS is disabled. This can be achieved with ether running (from the root)

```bash
nix develop
nix run ./#kontrol-service -- -dev-mode
```

Or using the helper script in the `kontrol-service` directory.

```bash
nix develop
cd kontrol-service
./dev-start-kk.sh
```

## Updating dependencies

### bun dependencies

To install an npm dependency with `bun`, simply use `bun i --save
some-package`. Commit the resulting changes to `package.json` and `bun.lockb`.
If there is no diff on `bun.lockb` but there is a diff on `package.json` after
installing, try running `bun i` again.

After installing a new dependency with `bun`, you must bump the version number
in `pin.json` to tell nix to build a new derivation of the dependencies. After
bumping this version number, you will need to build the package in order to get
the new derivation hash to add to `pin.json`.

```bash
nix develop
nix build ./#kontrol-frontend
```

This will output an error like
```
error: hash mismatch in fixed-output derivation '/nix/store/ri3g5rr26pkp29syvp1hb4v52c2m9y9z-kontrol-frontend-node_modules-0.7.0.drv':
         specified: <some hash>
            got:    <some hash>
```

Put the new `got` hash under `aarch64-darwin` in `pin.json`.
At this point you can push to a remote branch and run CI in order to get the
derivation hash for `x86_64-linux` and update `pin.json` with that hash as
well. Other architectures can be ignored for now.

### nix dependencies

`kontrol-frontend` depends on the `cli-kontrol-api` nix package for
auto-generated TypeScript types. This package is built from the main [kardinal
repo](https://github.com/kurtosis-tech/kardinal). The commit hash for this is
specified in the top level `flake.nix` file under `kardinal.url`. Once this is
updated, you can re-run `nix develop` and you should see an update to
`flake.lock` and possibly to `kontrol-frontend/bun.lockb` as well. Commit all
of these diffs.

