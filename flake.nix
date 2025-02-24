{
  description = "Nix flake for ocm";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, ... }:
    let
      pname = "ocm";

      # System types to support.
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import (builtins.fetchGit {
        # https://lazamar.co.uk/nix-versions/?package=go&channel=nixpkgs-unstable
        # https://lazamar.co.uk/nix-versions/?package=go&version=1.24rc2&fullName=go-1.24rc2&keyName=go_1_24&revision=21808d22b1cda1898b71cf1a1beb524a97add2c4&channel=nixpkgs-unstable
        name = "1.24.0";
        url = "https://github.com/NixOS/nixpkgs.git";
        ref = "refs/heads/nixos-unstable";
        # take latest commit sha from https://github.com/NixOS/nixpkgs/commits/nixos-unstable/
        rev = "73cf49b8ad837ade2de76f87eb53fc85ed5d4680";
      }) { inherit system; });

    in
    {
      # Provide some binary packages for selected system types.
      packages = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
          inherit (pkgs) stdenv lib ;
        in
        {
          ${pname} = pkgs.buildGoModule.override { go = pkgs.go_1_24; } rec {
            inherit pname self;
            version = lib.fileContents ./VERSION;
            gitCommit = if (self ? rev) then self.rev else self.dirtyRev;
            state = if (self ? rev) then "clean" else "dirty";

            # This vendorHash represents a derivative of all go.mod dependencies and needs to be adjusted with every change
            vendorHash = "sha256-aKKfiwYk6cKynjSXuQLCqeyLUCAZ4UavDXC72KFbItc=";

            src = ./.;

            ldflags = [
              "-s" "-w"
              "-X ocm.software/ocm/api/version.gitVersion=${version}"
              "-X ocm.software/ocm/api/version.gitTreeState=${state}"
              "-X ocm.software/ocm/api/version.gitCommit=${gitCommit}"
            # "-X ocm.software/ocm/api/version.buildDate=1970-01-01T0:00:00+0000"
            ];

            CGO_ENABLED = 0;

            subPackages = [
              "cmds/ocm"
              "cmds/cliplugin"
              "cmds/demoplugin"
              "cmds/ecrplugin"
              "cmds/helminstaller"
              "cmds/subcmdplugin"
              "cmds/jfrogplugin"
            ];

            nativeBuildInputs = [ pkgs.installShellFiles ];

            postInstall = ''
              installShellCompletion --cmd ${pname} \
                  --zsh  <($out/bin/${pname} completion zsh) \
                  --bash <($out/bin/${pname} completion bash) \
                  --fish <($out/bin/${pname} completion fish)
            '';

            meta = with lib; {
              description = "Open Component Model (OCM) is an open standard to describe software bills of delivery (SBOD)";
              longDescription = ''
                OCM is a technology-agnostic and machine-readable format focused on the software artifacts that must be delivered for software products.
                The specification is also used to express metadata needed for security, compliance, and certification purpose.
              '';
              homepage = "https://ocm.software";
              license = licenses.asl20;
              platforms = supportedSystems;
            };
          };
        });

      # Add dependencies that are only needed for development
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go_1_24   # golang 1.24
              gopls     # go language server
              gotools   # go imports
              go-tools  # static checks
              gnumake   # standard make
            ];
          };
        });

      # The default package for 'nix build'.
      defaultPackage = forAllSystems (system: self.packages.${system}.${pname});

      # These are the apps included in the default package.
      apps = forAllSystems (system: rec {
        ${pname} = default;
        default = {
            type = "app";
            program = self.packages.${system}.${pname} + "/bin/ocm";
        };
        helminstaller = {
          type = "app";
          program = self.packages.${system}.${pname} + "/bin/helminstaller";
        };
        demo = {
          type = "app";
          program = self.packages.${system}.${pname} + "/bin/demoplugin";
        };
        ecrplugin = {
          type = "app";
          program = self.packages.${system}.${pname} + "/bin/ecrplugin";
        };
      });

      legacyPackages = forAllSystems (system: rec {
        nixpkgs = nixpkgsFor.${system};
      });

    };
}
