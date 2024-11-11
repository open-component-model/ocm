{
  description = "Nix flake for ocm";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    # nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05"; # doesn
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
          ${pname} = pkgs.buildGoModule.override { go = pkgs.go_1_23; } rec {
            inherit pname self;
            version = lib.fileContents ./VERSION;
            gitCommit = if (self ? rev) then self.rev else self.dirtyRev;
            state = if (self ? rev) then "clean" else "dirty";

            # This vendorHash represents a derivative of all go.mod dependencies and needs to be adjusted with every change
            vendorHash = "sha256-a4es06aTjdnSXV9Xseh3WZ6zprWGM/Yk2PNhQwUZnxM=";

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
              "cmds/helminstaller"
              "cmds/demoplugin"
              "cmds/ecrplugin"
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
              go_1_23   # golang 1.23
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
