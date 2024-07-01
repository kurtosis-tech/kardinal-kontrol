{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
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

        service_names = ["kontrol-service" "kontrol-frontend"];
        architectures = ["amd64" "arm64"];
        imageRegistry = "258623609258.dkr.ecr.us-east-1.amazonaws.com/kurtosistech";

        matchingContainerArch =
          if builtins.match "aarch64-.*" system != null
          then "arm64"
          else if builtins.match "x86_64-.*" system != null
          then "amd64"
          else throw "Unsupported system type: ${system}";

        mergeContainerPackages = acc: service:
          pkgs.lib.recursiveUpdate acc {
            packages."${service}-container" = self.containers.${system}.${service}.${matchingContainerArch};
          };

        multiPlatformDockerPusher = acc: service:
          pkgs.lib.recursiveUpdate acc {
            packages."publish-${service}-container" = let
              name = "${imageRegistry}/${service}";
              tagBase = "latest";
              images =
                map (
                  arch: rec {
                    inherit arch;
                    image = self.containers.${system}.${service}.${arch};
                    tag = "${tagBase}-${arch}";
                  }
                )
                architectures;
              loadAndPush = builtins.concatStringsSep "\n" (pkgs.lib.concatMap
                ({
                  arch,
                  image,
                  tag,
                }: [
                  "$docker load -i ${image}"
                  "$docker tag ${service} ${name}:${tag}"
                  "$docker push ${name}:${tag}"
                ])
                images);
              imageNames =
                builtins.concatStringsSep " "
                (map ({
                  arch,
                  image,
                  tag,
                }: "${name}:${tag}")
                images);
            in
              pkgs.writeTextFile {
                inherit name;
                text = ''
                  #!${pkgs.stdenv.shell}
                  set -euxo pipefail
                  docker=${pkgs.docker}/bin/docker
                  ${loadAndPush}
                  $docker manifest create --amend ${name}:${tagBase} ${imageNames}
                  $docker manifest push ${name}:${tagBase}
                '';
                executable = true;
                destination = "/bin/push";
              };
          };

        mkGoApplicationImage = {
          pkgs,
          container_pkgs,
          service_name,
          service,
          arch,
          os,
          needsCrossCompilation,
        }: let
          overrideService =
            if !needsCrossCompilation
            then
              service.overrideAttrs
              (old: old // {doCheck = false;})
            else
              service.overrideAttrs (old:
                old
                // {
                  GOOS = os;
                  GOARCH = arch;
                  # CGO_ENABLED = disabled breaks the CLI compilation
                  # CGO_ENABLED = 0;
                  doCheck = false;
                });
        in
          pkgs.dockerTools.buildImage {
            name = "${service_name}";
            tag = "latest-${arch}";
            # tag = commit_hash;
            created = "now";
            copyToRoot = pkgs.buildEnv {
              name = "image-root";
              paths = [
                overrideService
                container_pkgs.bashInteractive
                container_pkgs.nettools
                container_pkgs.gnugrep
                container_pkgs.coreutils
              ];
              pathsToLink = ["/bin"];
            };
            architecture = arch;
            config.Cmd =
              if !needsCrossCompilation
              then ["${overrideService}/bin/${overrideService.pname}"]
              else ["${overrideService}/bin/${os}_${arch}/${overrideService.pname}"];
          };

        mkFrontendImage = {
          pkgs,
          container_pkgs,
          ...
        }:
          pkgs.callPackage ./kontrol-frontend/image.nix {
            inherit container_pkgs pkgs;
          };

        systemOutput = rec {
          devShells.default = pkgs.callPackage ./shell.nix {
            inherit pkgs;
          };

          packages.kontrol-service = pkgs.callPackage ./kontrol-service/default.nix {
            inherit pkgs;
          };

          packages.kontrol-frontend = pkgs.callPackage ./kontrol-frontend/default.nix {
            inherit pkgs;
          };

          containers = let
            os = "linux";
            all =
              pkgs.lib.mapCartesianProduct ({
                arch,
                service_name,
              }: {
                "${service_name}" = {
                  "${toString arch}" = let
                    nix_arch =
                      builtins.replaceStrings
                      ["arm64" "amd64"] ["aarch64" "x86_64"]
                      arch;

                    container_pkgs = import nixpkgs {
                      system = "${nix_arch}-${os}";
                    };

                    # if running from linux no cross-compilation is needed to palce the service in a container
                    needsCrossCompilation =
                      "${nix_arch}-${os}"
                      != system;
                  in
                    if service_name == "kontrol-frontend"
                    then mkFrontendImage {inherit pkgs container_pkgs;}
                    else
                      mkGoApplicationImage {
                        inherit pkgs container_pkgs service_name arch os needsCrossCompilation;
                        service = packages.${service_name};
                      };
                };
              }) {
                arch = architectures;
                service_name = service_names;
              };
          in
            pkgs.lib.foldl' (set: acc: pkgs.lib.recursiveUpdate acc set) {}
            all;
        };
        # Add containers matching architecture with local system as toplevel packages
        # this means calling `nix build .#<SERVICE_NAME>-container` will build the container matching the local system.
        # For cross-compilation use the containers attribute directly: `nix build .containers.<LOCAL_SYSTEM>.<SERVICE_NAME>.<ARCH>`
        outputWithContaniers = pkgs.lib.foldl' mergeContainerPackages systemOutput service_names;
        outputWithContainersAndPushers = pkgs.lib.foldl' multiPlatformDockerPusher outputWithContaniers service_names;
      in
        outputWithContainersAndPushers
    );
}
