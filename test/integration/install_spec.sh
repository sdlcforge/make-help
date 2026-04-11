# shellcheck shell=bash
# shellcheck disable=SC2317,SC2034

Describe 'scripts/install.js'
  EXPECTED_VERSION="$(node -p "require('$PWD/package.json').version")"

  setup_workdir() {
    WORKDIR="$(mktemp -d)"
    cp "$PWD/package.json" "$WORKDIR/"
    cp -r "$PWD/scripts"   "$WORKDIR/"
    cp "$PWD/go.mod"       "$WORKDIR/"
    cp "$PWD/go.sum"       "$WORKDIR/"
    cp -r "$PWD/cmd"       "$WORKDIR/"
    cp -r "$PWD/internal"  "$WORKDIR/"
  }

  cleanup_workdir() {
    if [ -n "${WORKDIR:-}" ] && [ -d "$WORKDIR" ]; then
      rm -rf "$WORKDIR"
    fi
  }

  Describe 'download pre-built binary'
    Before 'setup_workdir'
    After 'cleanup_workdir'

    It 'installs the binary successfully'
      When run node "$WORKDIR/scripts/install.js"
      The status should be success
      The output should include "Successfully installed"
    End

    It 'produces an executable binary'
      node "$WORKDIR/scripts/install.js" >/dev/null 2>&1
      When run test -x "$WORKDIR/bin/make-help"
      The status should be success
    End

    It 'binary reports the correct version'
      node "$WORKDIR/scripts/install.js" >/dev/null 2>&1
      When run "$WORKDIR/bin/make-help" --version
      The output should include "$EXPECTED_VERSION"
      The status should be success
    End
  End

  Describe '--build from source'
    go_missing() { ! command -v go >/dev/null 2>&1; }
    Skip if 'Go is not installed' go_missing
    Before 'setup_workdir'
    After 'cleanup_workdir'

    It 'builds the binary successfully'
      When run node "$WORKDIR/scripts/install.js" --build
      The status should be success
      The output should include "Successfully built"
    End

    It 'produces an executable binary'
      node "$WORKDIR/scripts/install.js" --build >/dev/null 2>&1
      When run test -x "$WORKDIR/bin/make-help"
      The status should be success
    End

    It 'binary reports the correct version'
      node "$WORKDIR/scripts/install.js" --build >/dev/null 2>&1
      When run "$WORKDIR/bin/make-help" --version
      The output should include "$EXPECTED_VERSION"
      The status should be success
    End
  End
End
