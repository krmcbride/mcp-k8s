{
  description = "mcp-k8s";

  inputs = {
    nixpkgs.url = "nixpkgs/8b49874f4334f6eb7bab76e564a462d336e5a1cb";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        go = pkgs.go_1_24;
      in {
        devShell = pkgs.mkShellNoCC{
          name = "mcp-k8s";
          nativeBuildInputs = with pkgs; [
            go
          ];
          CGO_ENABLED = 0;
        };
      }
    );
}
