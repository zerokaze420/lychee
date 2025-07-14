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
        packages.default = pkgs.buildGoModule {
          pname = "lychee"; # ✨ 我把名字改成了 lychee
          version = "0.1.0";
          src = ./.;
          # 👇 记得替换成真实的 hash
          vendorHash = "sha256-RIjhPcNyIISq7QF1k2aRyMzA5Eh/rv+epL5BZ+LmPCs=";

          # nativeBuildInputs = go-systemd-deps;
            nativeBuildInputs = [ pkgs.systemd pkgs.pkg-config ];
        };

        # ⭐️ 新增的部分：定义测试
        checks.default = pkgs.runCommand "go-unit-tests" {
          # 将构建依赖也作为测试的依赖
          nativeBuildInputs = [ pkgs.go ] ++ go-systemd-deps;
          src = ./.;
        } ''
          # 进入项目源码目录
          cd $src

          # 运行 Go 的标准测试命令
          # 如果测试失败，命令会以非零状态码退出，CI 就会失败
          go test ./...

          # 创建一个空的 $out 文件表示测试成功
          echo "Go tests passed" > $out
        '';

        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go
          ] ++ go-systemd-deps;
        };
      }
    );
}