{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem
    (
      system: let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
        with pkgs; {
          devShells.default = mkShell {
            buildInputs = [k3d kubectl kustomize argo-rollouts kubernetes-helm minikube];
            shellHook = ''
              source <(kubectl completion bash)
              source <(kubectl-argo-rollouts completion bash)
              source <(minikube completion bash)
            '';
          };
        }
    );
}
