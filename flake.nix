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
        # 👇 把依赖定义在这里，方便复用
        go-systemd-deps = [
          pkgs.systemd      # 提供 .h 头文件和 .so 库文件
          pkgs.pkg-config   # CGO 用来查找库的工具
        ];
      in
      {
        # --- 1. 如果你是为了最终打包（例如构建 Docker 镜像或二进制文件）---
        packages.default = pkgs.buildGoModule {
          pname = "my-go-app";
          version = "0.1.0";
          src = ./.;
          vendorHash = pkgs.lib.fakeSha256; # 替换成你的 vendorHash

          # CGO 需要的构建工具
          nativeBuildInputs = go-systemd-deps;
        };

        # --- 2. 如果你是为了开发环境（nix develop）---
        devShells.default = pkgs.mkShell {
          # 开发环境中需要的工具和库
          packages = [
            pkgs.go
          ] ++ go-systemd-deps; # 直接将依赖加入
        };
      }
    );
}