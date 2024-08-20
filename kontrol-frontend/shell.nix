{pkgs, ...}:
pkgs.mkShell {
  buildInputs = [
    pkgs.bun
    pkgs.nodejs
    pkgs.git
  ];
  shellHook = ''
    ROOT_DIR=$(git rev-parse --show-toplevel)
    ln -sfn ${pkgs.kardinal.cli-kontrol-api} $ROOT_DIR/.cli-kontrol-api
    bun i # install dependencies
  '';
}
