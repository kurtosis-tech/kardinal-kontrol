{
  pkgs ? (let
    inherit (builtins) fetchTree fromJSON readFile;
    inherit ((fromJSON (readFile ../flake.lock)).nodes) nixpkgs gomod2nix;
  in
    import (fetchTree nixpkgs.locked) {
      overlays = [(import "${fetchTree gomod2nix.locked}/overlay.nix")];
    }),
  mkGoEnv ? pkgs.mkGoEnv,
  gomod2nix ? pkgs.gomod2nix,
}: let
  goEnv = mkGoEnv {pwd = ./.;};
in
  pkgs.mkShell {
    nativeBuildInputs = with pkgs; [
      goEnv

      goreleaser
      go
      gopls
      golangci-lint
      delve
      enumer
      gomod2nix
      bash-completion
    ];
  }
