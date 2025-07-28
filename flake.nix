{
  description = "My Go application"; # 更新描述，不再提及 systemd

  inputs = {
    nixpkgs.url = "github.com/NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github.com/numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        # 移除了 pkgs.systemd
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
            # 移除了对 systemd 库路径的引用，如果你的 Go 代码不再使用 libsystemd，
            # 这行就不再需要了。如果 Go 代码仍然隐式地需要它，编译会失败。
            # export CGO_LDFLAGS="-Wl,-rpath,${pkgs.lib.makeLibraryPath [ pkgs.systemd ]}"
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
          ] ++ go-deps; # 使用更新后的 go-deps
        };
      }
    );
}
