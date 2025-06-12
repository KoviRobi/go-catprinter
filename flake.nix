{
  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
    }:
    flake-parts.lib.mkFlake { inherit inputs; } (
      {
        config,
        withSystem,
        moduleWithSystem,
        ...
      }:
      {
        systems = [
          "x86_64-linux"
          "aarch64-linux"
          "x86_64-darwin"
          "aarch64-darwin"
        ];
        perSystem =
          {
            config,
            lib,
            pkgs,
            ...
          }:
          {
            packages.default = pkgs.callPackage ./derivation.nix {
              version = self.rev or "unstable${builtins.substring 0 8 self.lastModifiedDate}";
              src = lib.cleanSource self;
            };
          };
      }
    );
}
