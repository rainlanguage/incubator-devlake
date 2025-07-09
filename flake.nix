{
  description = "Nix flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        pkgs = pkgs;

        devShells.default = pkgs.mkShell {
          nativeBuildInputs = [ pkgs.go pkgs.gopls ];
          shellHook = 
            # there is a known issue with nix pkgs new apple_sdk and that since it is now using xcrun,
            # apple_sdk's setup hook breaks the link to some of '/usr/bin' Xcode command line tools bins
            # and libs, this mainly is an issue for `tauri-shell` devshell when tauri cli is used to build
            # a tauri app for macos where it needs one of those bins called `SetFile`, for more details:
            # https://github.com/NixOS/nixpkgs/issues/355486
            #
            # this is a workaround that removes xcrun from devshell PATH and unsets DEVELOPER_DIR so that
            # those apple bins and libs are accessible normally through `/usr/bin`
            (if pkgs.stdenv.isDarwin then ''
              export PATH=''${PATH//'${pkgs.xcbuild.xcrun}/bin:'/}
              unset DEVELOPER_DIR
            '' else
              "");
        };
      });
}
