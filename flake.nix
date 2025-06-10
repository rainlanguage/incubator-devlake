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
        };
      });
}
