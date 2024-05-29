{
  pkgs ? (let
    inherit (builtins) fetchTree fromJSON readFile;
    inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
  in
    import (fetchTree nixpkgs.locked) {
      overlays = [(import "${fetchTree gomod2nix.locked}/overlay.nix")];
    }),
  buildGoApplication ? pkgs.buildGoApplication,
  commit_hash ? "dirty",
}: let
  pname = "kontrol-service";
  ldflags = pkgs.lib.concatStringsSep "\n" [
    "-X github.com/kurtosis-tech/kurtosis/kardinal.AppName=${pname}"
    "-X github.com/kurtosis-tech/kurtosis/kardinal.Commit=${commit_hash}"
  ];
in
  buildGoApplication {
    # pname has to match the location (folder) where the main function is or use
    # subPackges to specify the file (e.g. subPackages = ["some/folder/main.go"];)
    inherit pname ldflags;
    pwd = ./.;
    src = ./.;
    modules = ./gomod2nix.toml;
    CGO_ENABLED = 0;
  }
