/*
SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors

SPDX-License-Identifier: Apache-2.0
*/
{
  description = "Nix flake for ocm";

  inputs = {
    # NixPkgs (nixos-23.11)
    nixpkgs.url = "nixpkgs/nixos-23.11"; #"github:NixOS/nixpkgs/nixos-23.11";
  };

  outputs = { self, nixpkgs, ... }:
    let
      pname = "ocm";

      # System types to support.
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });

    in
    {
      # Provide some binary packages for selected system types.
      packages = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
          inherit (pkgs) stdenv lib ;
        in
        {
          ${pname} = pkgs.buildGo121Module rec {
            inherit pname self;
            version = lib.fileContents ./VERSION;
            gitCommit = if (self ? rev) then self.rev else self.dirtyRev;
            state = if (self ? rev) then "clean" else "dirty";

            # This vendorHash represents a dervative of all go.mod dependancies and needs to be adjusted with every change
            vendorHash = "sha256-CA3p9QNHo7mHCoXkuOojFAJen3TdieSsoQVivxPk3yw=";

            src = ./.;

            ldflags = [
              "-s" "-w"
              "-X github.com/open-component-model/ocm/pkg/version.gitVersion=${version}"
              "-X github.com/open-component-model/ocm/pkg/version.gitTreeState=${state}"
              "-X github.com/open-component-model/ocm/pkg/version.gitCommit=${gitCommit}"
            # "-X github.com/open-component-model/ocm/pkg/version.buildDate=1970-01-01T0:00:00+0000"
            ];

            CGO_ENABLED = 0;

            subPackages = [ 
              "cmds/ocm" 
              "cmds/helminstaller"
              "cmds/demoplugin"
              "cmds/ecrplugin"
            ];

            nativeBuildInputs = [ pkgs.installShellFiles ];

            postInstall = ''
              installShellCompletion --cmd ${pname} --zsh  <($out/bin/${pname} completion zsh)
              installShellCompletion --cmd ${pname} --bash <($out/bin/${pname} completion bash)
              installShellCompletion --cmd ${pname} --fish <($out/bin/${pname} completion fish)
            '';

            meta = with lib; {
              description = "Open Component Model (OCM) is an open standard to describe software bills of delivery (SBOD)";
              longDescription = ''
                The Open Component Model (OCM) is an open standard to describe software bills of delivery (SBOD). OCM is a technology-agnostic and machine-readable format focused on the software artifacts that must be delivered for software products.
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
              go_1_21 
              gopls 
              gotools 
              go-tools
              gnumake
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
    };
}
