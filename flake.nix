{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix.url = "github:nix-community/gomod2nix";
    gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
    gomod2nix.inputs.flake-utils.follows = "flake-utils";
    kardinal.url = "github:kurtosis-tech/kardinal/a1632d5aecd8";
  };
  outputs = {
    self,
    nixpkgs,
    flake-utils,
    gomod2nix,
    kardinal,
    ...
  }:
    flake-utils.lib.eachDefaultSystem
    (
      system: let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [
            (import "${gomod2nix}/overlay.nix")
            (final: prev: {
              kardinal = kardinal.outputs.packages.${system};
            })
          ];
        };

        service_names = ["kontrol-service" "kontrol-frontend"];
        architectures = ["amd64" "arm64"];
        imageRegistries = {
          "aws" = "258623609258.dkr.ecr.us-east-1.amazonaws.com";
          "dockerhub" = "docker.io";
        };

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

        mergeKardinalPackages = acc: packageName: package:
          pkgs.lib.recursiveUpdate acc {
            packages."public-${packageName}" = package;
          };

        multiPlatformDockerPusher = acc: service: let
          scriptBuilder = imageRegistry: let
            name = "${imageRegistry}/kurtosistech/${service}";
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
                "$docker tag ${service}:${tag} ${name}:${tag}"
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
          packages =
            pkgs.lib.concatMapAttrs (name: registry: {
              "publish-${name}-${service}-container" = scriptBuilder registry;
            })
            imageRegistries;
        in
          pkgs.lib.recursiveUpdate acc {packages = packages;};

        getPkgsCrossSystem = targetSystem:
          builtins.trace "${targetSystem}" (
            if "aarch64-darwin" == targetSystem
            then pkgs.pkgsCross.${targetSystem}
            else if "x86_64-darwin" == targetSystem
            then pkgs.pkgsCross.${targetSystem}
            else if "aarch64-linux" == targetSystem
            then pkgs.pkgsCross.aarch64-multiplatform
            else if "x86_64-linux" == targetSystem
            then pkgs.pkgsCross.gnu64
            else throw "Unsupported targetSystem type: ${targetSystem}"
          );

        mkGoApplicationImage = {
          pkgs,
          container_pkgs,
          service_name,
          arch,
          os,
          needsCrossCompilation,
        }: let
          targetSystem = container_pkgs.stdenv.hostPlatform.system;
          targetCrossPkgs = getPkgsCrossSystem targetSystem;
          serviceCross = import (./. + builtins.toPath "/${service_name}/default.nix") {pkgs = targetCrossPkgs;};
          serviceNative = import (./. + builtins.toPath "/${service_name}/default.nix") {inherit pkgs;};
          overrideService =
            if !needsCrossCompilation
            then
              serviceNative.overrideAttrs
              (old: old // {doCheck = false;})
            else
              serviceCross.overrideAttrs (old:
                old
                // {
                  GOOS = os;
                  GOARCH = arch;
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
                container_pkgs.cacert
                container_pkgs.python3 # Add Python 3 to the container for plugins
                container_pkgs.git # Add Git to the container for plugins
              ];
              pathsToLink = ["/bin"];
            };
            architecture = arch;
            config.Cmd =
              if !needsCrossCompilation
              then ["${overrideService}/bin/${overrideService.pname}"]
              else ["${overrideService}/bin/${os}_${arch}/${overrideService.pname}"];
            config.Env = [
              "SSL_CERT_FILE=${container_pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
              "PATH=${container_pkgs.python3}/bin:${container_pkgs.git}/bin:$PATH"
            ];
          };

        mkFrontendImage = {
          pkgs,
          container_pkgs,
          arch,
          ...
        }:
          pkgs.callPackage ./kontrol-frontend/image.nix {
            inherit container_pkgs pkgs;
            tag = "latest-${arch}";
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

          packages.default = packages.kontrol-service;

          # Bypass public kardinal definitions
          kardinal = pkgs.kardinal;

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
                    then
                      mkFrontendImage {
                        inherit pkgs container_pkgs arch;
                      }
                    else
                      mkGoApplicationImage {
                        inherit pkgs container_pkgs service_name arch os needsCrossCompilation;
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
        outputWithContainersAndPushersAndKardinal = pkgs.lib.attrsets.foldlAttrs mergeKardinalPackages outputWithContainersAndPushers pkgs.kardinal;
      in
        outputWithContainersAndPushersAndKardinal
    );
}
