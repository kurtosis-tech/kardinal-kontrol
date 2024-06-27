{
  pkgs,
  container_pkgs,
  ...
}: let
  name = "kontrol-frontend";
  nginxWebRoot = import ./default.nix {inherit pkgs;};
  nginxPort = "80";

  nginxConf = pkgs.writeText "nginx.conf" ''
    user nobody nobody;
    daemon off;
    error_log /dev/stdout info;
    pid /dev/null;
    events {
      worker_connections 1024;
    }
    http {
      include       ${container_pkgs.nginx}/conf/mime.types;
      default_type  application/octet-stream;

      add_header Cache-Control "no-store";

      access_log /dev/stdout;
      server {
        listen ${nginxPort};
        server_name localhost;
        proxy_cache off;

        root ${nginxWebRoot};
        index index.html;

        location / {
            try_files $uri $uri/ /index.html =404;
        }

        error_page 404 /404.html;
        error_page 500 502 503 504 /50x.html;
      }
    }
  '';
in
  pkgs.dockerTools.buildLayeredImage {
    inherit name;
    tag = "latest";

    contents = [
      container_pkgs.nginx
      container_pkgs.fakeNss
    ];

    extraCommands = ''
      mkdir -p var/log/nginx
      mkdir -p var/cache/nginx
      mkdir -p tmp
      chmod 1777 tmp
    '';

    config = {
      Cmd = ["nginx" "-c" nginxConf];
      ExposedPorts = {
        "${nginxPort}/tcp" = {};
      };
    };
  }
