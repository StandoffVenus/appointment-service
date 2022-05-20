{ pkgs ? import <nixpkgs> { } }:

with pkgs;
let
  go = go_1_18;
in mkShell {
  buildInputs = [ go sqlite ];
}
