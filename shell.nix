let
  unstableTarball =
    fetchTarball https://github.com/NixOS/nixpkgs/archive/nixos-unstable.tar.gz;
  unstable = import unstableTarball {};
  pkgs = unstable; # Use the unstable packages
in
pkgs.mkShell {
  buildInputs = [
    unstable.go_1_23 # You can change the version as needed (e.g., go_1_20)
    pkgs.gotools # For various Go tools
    pkgs.golint # For code linting
    pkgs.gofumpt # For code formatting
    pkgs.git # For version control
  ];
  # Optional, but useful for consistency
  GO111MODULE = "on";
  GOPROXY = "https://proxy.golang.org,direct";
}