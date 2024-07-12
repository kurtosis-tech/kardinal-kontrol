{pkgs}:
# Based on the https://github.com/NixOS/nixpkgs/blob/master/pkgs/by-name/he/helix-gpt/package.nix
# for the Bun compilation and bundling
with pkgs; let
  pin = lib.importJSON ./pin.json;
  name = "kontrol-frontend";
  src = ./.;
  node_modules = stdenv.mkDerivation {
    pname = "${name}-node_modules";
    inherit src;
    version = pin.version;
    impureEnvVars =
      lib.fetchers.proxyImpureEnvVars
      ++ ["GIT_PROXY_COMMAND" "SOCKS_SERVER"];
    nativeBuildInputs = [bun];
    dontConfigure = true;

    # Dont try to patch shebangs in shell scripts contained in vendored node_modules packages
    # https://nixos.org/manual/nixpkgs/stable/#var-stdenv-dontPatchShebangs
    #
    # And if you run into issues with /nix/store/ paths leaking into the fixed output of this derivation
    # run the build with the following command:
    # > nix build .#kontrol-frontend --no-link --print-out-paths --print-build-logs
    # The flag --print-build-logs will show you the files that leaking the paths.
    dontPatchShebangs = true;

    buildPhase = ''
      # Mimic local of develop copy of kardinal
      cp -r ${pkgs.kardinal.cli-kontrol-api} ../.cli-kontrol-api
      bun install --no-progress --frozen-lockfile
    '';

    installPhase = ''
      mkdir -p $out/node_modules
      cp -R ./node_modules $out
    '';

    outputHash = pin."${stdenv.system}";
    outputHashAlgo = "sha256";
    outputHashMode = "recursive";
  };
in
  stdenv.mkDerivation {
    pname = "${name}";
    version = pin.version;
    inherit src;
    nativeBuildInputs = [makeBinaryWrapper rsync];

    dontConfigure = true;

    buildPhase = ''
      runHook preBuild

      # Because bun link of local deps (file:...) the link from the previous derivation are broken
      # we copy all but the those modules and re-copying them directly
      mkdir -p node_modules
      ln -sfn ${pkgs.kardinal.cli-kontrol-api} node_modules/cli-kontrol-api
      rsync -av --progress ${node_modules}/node_modules . --exclude cli-kontrol-api

      ${bun}/bin/bun run --no-install build

      runHook postInstall
    '';

    installPhase = ''
      runHook preInstall
      cp -R ./dist $out
      runHook postInstall
    '';
  }
