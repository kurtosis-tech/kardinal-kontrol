{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    flake-utils.url = "github:numtide/flake-utils";
    unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
    gomod2nix.url = "github:nix-community/gomod2nix";
    gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
    gomod2nix.inputs.flake-utils.follows = "flake-utils";
  };
  outputs = {
    self,
    nixpkgs,
    flake-utils,
    unstable,
    gomod2nix,
    ...
  }:
    flake-utils.lib.eachDefaultSystem
    (
      system: let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [
            (import "${gomod2nix}/overlay.nix")
          ];
        };
        commit_hash = "${self.shortRev or self.dirtyShortRev or "dirty"}";
        service_names = ["kardinal-manager" "redis-proxy-overlay"];
      in rec {
        devShells.default = pkgs.callPackage ./shell.nix {
          inherit pkgs;
        };

        packages.kardinal-manager = pkgs.callPackage ./kontrol-service/default.nix {
          inherit pkgs;
        };

        packages.redis-proxy-overlay = pkgs.callPackage ./redis-overlay-service/default.nix {
          inherit pkgs;
        };

        packages.kontrol-frontend = pkgs.callPackage ./kontrol-frontend/default.nix {
          inherit pkgs;
        };

        containers = let
          architectures = ["amd64" "arm64"];
          os = "linux";
          all = pkgs.lib.lists.crossLists (arch: service_name: {
            "${service_name}" = {
              "${toString arch}" = let
                # if running from linux no cross-compilation is needed to palce the service in a container
                needsCrossCompilation =
                  "${arch}-${os}"
                  != builtins.replaceStrings ["aarch64" "x86_64"] [
                    "arm64"
                    "amd64"
                  ]
                  system;
                service =
                  if !needsCrossCompilation
                  then
                    packages.${service_name}.overrideAttrs
                    (old: old // {doCheck = false;})
                  else
                    packages.${service_name}.overrideAttrs (old:
                      old
                      // {
                        GOOS = os;
                        GOARCH = arch;
                        # CGO_ENABLED = disabled breaks the CLI compilation
                        # CGO_ENABLED = 0;
                        doCheck = false;
                      });
              in
                builtins.trace "${service}/bin" pkgs.dockerTools.buildImage {
                  name = "kurtosistech/${service_name}";
                  tag = "latest";
                  # tag = commit_hash;
                  created = "now";
                  copyToRoot = pkgs.buildEnv {
                    name = "image-root";
                    paths = [
                      service
                      pkgs.bashInteractive
                      pkgs.nettools
                      pkgs.gnugrep
                    ];
                    pathsToLink = ["/bin"];
                  };
                  architecture = arch;
                  config.Cmd =
                    if !needsCrossCompilation
                    then ["${service}/bin/${service.pname}"]
                    else ["${service}/bin/${os}_${arch}/${service.pname}"];
                };
            };
          }) [architectures service_names];
        in
          pkgs.lib.foldl' (set: acc: pkgs.lib.recursiveUpdate acc set) {}
          all;

        packages.container = let
          arch =
            if builtins.match "aarch64-.*" system != null
            then "arm64"
            else if builtins.match "x86_64-.*" system != null
            then "amd64"
            else throw "Unsupported system type: ${system}";
          mergeCurrentPackage = acc: service: pkgs.lib.recursiveUpdate acc {"${service}" = self.containers.${system}.${service}.${arch};};
        in
          pkgs.lib.foldl' mergeCurrentPackage {} service_names;
      }
    );
}
