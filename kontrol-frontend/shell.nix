{ pkgs, ... }:

pkgs.mkShell {
  buildInputs = [
    pkgs.bun
  ];
}
