{pkgs}: let
  mergeShells = shells:
    pkgs.mkShell {
      shellHook = builtins.concatStringsSep "\n" (map (s: s.shellHook or "") shells);
      buildInputs = builtins.concatLists (map (s: s.buildInputs or []) shells);
      nativeBuildInputs = builtins.concatLists (map (s: s.nativeBuildInputs or []) shells);
      paths = builtins.concatLists (map (s: s.paths or []) shells);
    };
  frontend_shell = pkgs.callPackage ./kontrol-frontend/shell.nix {inherit pkgs;};
  backend_shell = pkgs.callPackage ./kontrol-service/shell.nix {inherit pkgs;};
  kardinal_shell = with pkgs;
    pkgs.mkShell {
      buildInputs = [awscli dive kubectl kustomize kubernetes-helm minikube istioctl tilt reflex];
      shellHook = ''
        source <(kubectl completion bash)
        source <(minikube completion bash)
        printf '\u001b[31m

                                          :::::
                                           :::::::
                                           ::   :::
                                          :::     ::
                                          ::   ::- :::
                                        :::         :::
                                       ::: :::    :::
                                     :::    ::    ::
                                   :::      ::   :::
                                 :::       :::   ::
                               :::        ::     ::
                            ::::       ::::     ::
                          ::::      ::::      :::
                       ::::::::::::::       ::::
                                       ::::::
                   :::::::::::::::::::::
               ::::::
            :::::
          :::



        \u001b[0m
        Starting Kardinal dev shell.
        \e[32m
        \e[0m
        '
      '';
    };
in
  mergeShells [backend_shell frontend_shell kardinal_shell]
