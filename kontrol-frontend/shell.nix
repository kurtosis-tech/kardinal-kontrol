{pkgs, ...}:
pkgs.mkShell {
  buildInputs = [
    pkgs.bun
    pkgs.nodejs
  ];
}
