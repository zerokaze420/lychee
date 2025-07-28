{
  description = "My Go application"; # 更新描述，不再提及 systemd

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    # 修正 flake-utils 的 URL，加上 https://
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        go-deps = [
          pkgs.pkg-config
        ];
      in
      {
        packages.default = pkgs.stdenv.mkDerivation {
          pname = "lychee";
          version = "0.1.0";
          src = ./.;

          nativeBuildInputs = [
            pkgs.go_1_24
            pkgs.pkg-config
            # 移除了 pkgs.systemd
          ];

          buildPhase = ''
            export HOME=$(pwd)
            export GOPROXY=https://goproxy.cn,direct
            go build -mod=vendor -v -o lychee ./cmd/app/main.go
          '';

          installPhase = ''
            mkdir -p $out/bin
            mv lychee $out/bin/
          '';
        };
        packages.lychee-Image =  pkgs.dockerTools.buildImage {
          name = "lychee";
          config = {
            Cmd = [ "/lychee" ];
            WorkingDir = "/app";
            Env = [
              "GO111MODULE=on"
              "GOPROXY=https://goproxy.cn,direct"
            ];
            Volumes = {};
          };
        };

        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go_1_24
          ] ++ go-deps; # 使用更新后的 go-deps
        };
      }
    );
}
