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
        # --- 1. 使用 mkDerivation 进行手动、精确的构建 ---
        packages.default = pkgs.stdenv.mkDerivation {
          pname = "lychee"; # 我把名字改成了你的项目名
          version = "0.1.0";
          src = ./.;

          # 在这里，我们需要明确列出所有构建工具，包括 Go 本身
          nativeBuildInputs = [
            pkgs.go
            pkgs.pkg-config
            pkgs.systemd
          ];

          # 我们完全重写构建和安装阶段，来精确执行我们想要的命令
          buildPhase = ''
            # ✅ 修复：设置一个可写的 HOME 目录，防止 /homeless-shelter 权限错误
            export HOME=$(pwd)

            # 设置 CGO_LDFLAGS，让最终的二进制文件知道在运行时去哪里找 systemd 的 .so 动态库文件
            export CGO_LDFLAGS="-rpath ${pkgs.lib.makeLibraryPath [ pkgs.systemd ]}"

            # 手动运行你熟悉的 Go build 命令
            # -v 参数可以显示详细的编译输出，方便调试
            go build -v -o lychee ./cmd/app/main.go
          '';

          installPhase = ''
            # 创建目标目录并把编译好的文件放进去
            mkdir -p $out/bin
            mv lychee $out/bin/
          '';
        };

        # --- 2. 开发环境部分保持不变 ---
        devShells.default = pkgs.mkShell {
          # 开发环境中需要的工具和库
          packages = [
            pkgs.go
          ] ++ go-systemd-deps; # 直接将依赖加入
        };
      }
    );
}