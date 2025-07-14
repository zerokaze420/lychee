{
  description = "My Go application that uses systemd";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        go-systemd-deps = [
          pkgs.systemd
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
            pkgs.systemd
          ];

          buildPhase = ''
            export HOME=$(pwd)
            export GOPROXY=https://goproxy.cn,direct
            export CGO_LDFLAGS="-Wl,-rpath,${pkgs.lib.makeLibraryPath [ pkgs.systemd ]}"
            go build -mod=vendor -v -o lychee ./cmd/app/main.go
          '';

          installPhase = ''
            mkdir -p $out/bin
            mv lychee $out/bin/
          '';
        };

        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go_1_24
          ] ++ go-systemd-deps;
        };
      }
    );
}