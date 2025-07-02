{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    devenv.url = "github:cachix/devenv";
    flake-utils.url = "github:numtide/flake-utils";
  };

  nixConfig = {
    extra-trusted-public-keys = "devenv.cachix.org-1:w1cLUi8dv3hnoSPGAuibQv+f9TZLr6cv/Hm9XgU50cw=";
    extra-substituters = "https://devenv.cachix.org";
  };

  outputs = {
    self,
    nixpkgs,
    devenv,
    flake-utils,
    ...
  } @ inputs: let
    pname = "avalanche-api";
  in flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        packages = {
          default = pkgs.buildGo124Module {
            inherit pname;
            version = self.shortRev or "${self.lastModifiedDate}-dirty";
            src = self;
            vendorHash = "sha256-X4esiZW9Yi4uRDxr1FkteB5gJ0ldh8V4d+nfWlhguew=";
            doCheck = false;
          };
        };
        devShells.${system}.default = devenv.lib.mkShell {
          inherit inputs pkgs;
          modules = [
            {
              scripts.go-test.exec = ''
                ${pkgs.gotestsum}/bin/gotestsum -f testname ./... $@
              '';
              scripts.go-run.exec = ''
                go run . $@
              '';
              packages = [pkgs.alejandra pkgs.cobra-cli pkgs.go pkgs.go-tools pkgs.golangci-lint pkgs.gotestsum pkgs.go-mod-upgrade pkgs.graphviz ];
              difftastic.enable = true;
              languages.go.enable = true;
              env = {
                COMPOSE_PROFILES = "base,postgres";
                COMPOSE_REMOVE_ORPHANS = true;
              };
              git-hooks.hooks = {
                alejandra.enable = true;
                deadnix.enable = true;
                gofmt.enable = true;
                govet.enable = true;
                staticcheck.enable = true;
                golangci-lint.enable = true;
              };
            }
          ];
        };
      }
    );
}
