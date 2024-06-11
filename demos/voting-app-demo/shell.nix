{pkgs, ...}: let
  pyEnv = pkgs.python3.buildEnv.override {
    extraLibs = [pkgs.python3Packages.click pkgs.python3Packages.requests];
    ignoreCollisions = true;
  };

  cli = pkgs.stdenv.mkDerivation {
    pname = "kardinal";
    version = "1.0.0";

    src = ./.;

    installPhase = ''
      mkdir -p $out/bin
      cp -R ./manifests $out/bin/
      echo "#!${pyEnv}/bin/python3" > $out/bin/kardinal
      cat cli.py >> $out/bin/kardinal
      chmod +x $out/bin/kardinal
    '';
  };

  load-genarator = pkgs.stdenv.mkDerivation {
    pname = "load-generator";
    version = "1.0.0";

    src = ./.;

    installPhase = ''
      mkdir -p $out/bin
      echo "#!${pyEnv}/bin/python3" > $out/bin/load-generator
      cat load-generator.py >> $out/bin/load-generator
      chmod +x $out/bin/load-generator
    '';
  };
in
  pkgs.mkShell {
    buildInputs = [
      cli
      load-genarator
      pyEnv
    ];
  }
