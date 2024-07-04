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
    buildPhase = ''
      # Mimic local of develop copy of kardinal
      cp -r ${pkgs.kardinal.cli-kontrol-api} ../.cli-kontrol-api

      bun install --no-progress --frozen-lockfile
    '';
    installPhase = ''
      mkdir -p $out/node_modules

      cp -R ./node_modules $out

      # Remove dependencies that are leaking their own /nix/store/ paths the the fixed output of this derivation
      rm -rf $out/node_modules/lodash/flake.lock
      rm -rf $out/node_modules/cytoscape/.github
      rm -rf $out/node_modules/.cache
      rm -rf $out/node_modules/vscode-languageclient
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

      # bun is referenced naked in the package.json generated script
      mkdir -p ./bin
      makeBinaryWrapper ${bun}/bin/bun ./bin/${name} \
        --prefix PATH : ${lib.makeBinPath [bun]} \
        --add-flags "run  build"

      # Call wrapper to build the project
      ./bin/${name}

      runHook postInstall
    '';

    installPhase = ''
      runHook preInstall
      cp -R ./dist $out
      runHook postInstall
    '';
  }
