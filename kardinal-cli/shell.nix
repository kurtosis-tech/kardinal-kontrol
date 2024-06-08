{pkgs, ...}:
pkgs.mkShell {
  buildInputs = [
    pkgs.python3
    pkgs.python3Packages.virtualenv
  ];

  shellHook = ''
    pushd "$(git rev-parse --show-toplevel)/kardinal-cli"
    if [ ! -d ".venv" ]; then
      python -m venv .venv
    fi
    source .venv/bin/activate
    .venv/bin/pip install -r requirements.txt

    export PYTHONPATH=.
    alias kardinal="python $(pwd)/cli.py"
    popd
  '';
}
