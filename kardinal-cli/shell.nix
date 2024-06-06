{ pkgs, ... }:

pkgs.mkShell {
  buildInputs = [
    pkgs.python3
    pkgs.python3Packages.virtualenv
  ];

  shellHook = ''
    echo "Setting up Python environment..."
    cd `git rev-parse --show-toplevel`/kardinal-cli
    if [ ! -d ".venv" ]; then
      python -m venv .venv
      .venv/bin/pip install -r requirements.txt
    fi
    source .venv/bin/activate
    export PYTHONPATH=.
    alias kardinal="python ./cli.py"
  '';
}
