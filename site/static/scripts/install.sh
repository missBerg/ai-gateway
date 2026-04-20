#!/usr/bin/env bash
# shellcheck shell=bash
#
# Envoy AI Gateway — one-command installer
#
# Installs Envoy Gateway (with the AI Gateway values file), the AI Gateway
# CRDs, and the AI Gateway controller onto the Kubernetes cluster that your
# current kubeconfig points at.
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/site/static/scripts/install.sh | bash
#   curl -sSL https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/site/static/scripts/install.sh | bash -s -- --version v0.5.0 --yes
#   bash install.sh --version v0.5.0 --eg-version v0.0.0-latest
#
# Flags:
#   --version <ver>      AI Gateway version (default: prompt, fallback v0.5.0)
#   --eg-version <ver>   Envoy Gateway version (default: v0.0.0-latest)
#   -y, --yes            Skip prompts, accept defaults, non-interactive
#   -h, --help           Print this help
#
# Environment:
#   AIGW_VERSION         Same as --version
#   EG_VERSION           Same as --eg-version
#   ASSUME_YES=1         Same as --yes
#
# Exit codes:
#   0  success
#   1  precondition failed (missing tool, no cluster, user aborted)
#   2  install step failed

set -euo pipefail

# ---- Defaults ---------------------------------------------------------------

DEFAULT_AIGW_VERSION="v0.5.0"
DEFAULT_EG_VERSION="v0.0.0-latest"
EG_VALUES_URL="https://raw.githubusercontent.com/envoyproxy/ai-gateway/main/manifests/envoy-gateway-values.yaml"

AIGW_VERSION="${AIGW_VERSION:-}"
EG_VERSION="${EG_VERSION:-}"
ASSUME_YES="${ASSUME_YES:-}"

# ---- Pretty output ----------------------------------------------------------

if [[ -t 1 ]]; then
  C_RESET=$'\033[0m'
  C_BOLD=$'\033[1m'
  C_DIM=$'\033[2m'
  C_PURPLE=$'\033[38;5;135m'
  C_GREEN=$'\033[32m'
  C_RED=$'\033[31m'
  C_YELLOW=$'\033[33m'
else
  C_RESET=""
  C_BOLD=""
  C_DIM=""
  C_PURPLE=""
  C_GREEN=""
  C_RED=""
  C_YELLOW=""
fi

info()  { printf "%s==>%s %s\n" "$C_PURPLE" "$C_RESET" "$1"; }
ok()    { printf "%s✓%s %s\n" "$C_GREEN" "$C_RESET" "$1"; }
warn()  { printf "%s!%s %s\n" "$C_YELLOW" "$C_RESET" "$1" >&2; }
err()   { printf "%s✗%s %s\n" "$C_RED" "$C_RESET" "$1" >&2; }

print_help() {
  # Skip shebang and shellcheck directive; print the header block.
  sed -n '4,29p' "$0" 2>/dev/null | sed 's/^# \{0,1\}//'
}

# ---- Arg parsing ------------------------------------------------------------

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)         AIGW_VERSION="${2:-}"; shift 2 ;;
    --version=*)       AIGW_VERSION="${1#*=}"; shift ;;
    --eg-version)      EG_VERSION="${2:-}"; shift 2 ;;
    --eg-version=*)    EG_VERSION="${1#*=}"; shift ;;
    -y|--yes)          ASSUME_YES=1; shift ;;
    -h|--help)         print_help; exit 0 ;;
    *)                 err "Unknown argument: $1"; print_help; exit 1 ;;
  esac
done

# ---- Interactive helpers ---------------------------------------------------
#
# When piped via `curl | bash`, stdin is the pipe — not the user's terminal —
# so plain `read` would either block forever or consume the script itself.
# We read from /dev/tty when it's available; otherwise we fall back to the
# default silently.

can_prompt() {
  [[ -z "$ASSUME_YES" && -r /dev/tty ]]
}

prompt_default() {
  # $1 = prompt text, $2 = default value
  # Writes the chosen value to stdout.
  local prompt="$1" default="$2" answer=""
  if ! can_prompt; then
    printf "%s" "$default"
    return
  fi
  # Show prompt on stderr so stdout stays clean for capture
  printf "%s%s%s [%s%s%s]: " "$C_BOLD" "$prompt" "$C_RESET" "$C_DIM" "$default" "$C_RESET" >/dev/tty
  read -r answer </dev/tty || true
  printf "%s" "${answer:-$default}"
}

confirm() {
  # $1 = prompt text. Returns 0 on yes, 1 on no.
  local prompt="$1" answer=""
  if ! can_prompt; then
    return 0
  fi
  printf "%s%s%s [Y/n]: " "$C_BOLD" "$prompt" "$C_RESET" >/dev/tty
  read -r answer </dev/tty || true
  case "${answer,,}" in
    ""|y|yes) return 0 ;;
    *)        return 1 ;;
  esac
}

# ---- Preflight --------------------------------------------------------------

preflight() {
  info "Checking prerequisites…"
  local missing=0

  for tool in kubectl helm; do
    if ! command -v "$tool" >/dev/null 2>&1; then
      err "$tool not found in PATH"
      missing=1
    fi
  done
  [[ $missing -eq 0 ]] || { err "Install kubectl and helm, then re-run."; exit 1; }

  if ! kubectl cluster-info >/dev/null 2>&1; then
    err "Cannot reach a Kubernetes cluster with the current kubeconfig."
    err "Point kubectl at a running cluster, then re-run."
    exit 1
  fi

  # Warn if EG is already installed with a different values file.
  if kubectl get ns envoy-gateway-system >/dev/null 2>&1; then
    warn "Namespace envoy-gateway-system already exists."
    warn "The installer will upgrade the existing release — review your custom config first if you have one."
  fi

  ok "kubectl, helm, and cluster reachable"
}

# ---- Version resolution -----------------------------------------------------

resolve_versions() {
  info "Selecting versions…"
  if [[ -z "$AIGW_VERSION" ]]; then
    AIGW_VERSION="$(prompt_default "AI Gateway version" "$DEFAULT_AIGW_VERSION")"
  fi
  if [[ -z "$EG_VERSION" ]]; then
    EG_VERSION="$(prompt_default "Envoy Gateway version" "$DEFAULT_EG_VERSION")"
  fi
  ok "AI Gateway: $AIGW_VERSION"
  ok "Envoy Gateway: $EG_VERSION"
}

# ---- Plan summary + confirmation -------------------------------------------

summarize_and_confirm() {
  printf "\n"
  printf "%sThis script will run:%s\n" "$C_BOLD" "$C_RESET"
  printf "  %s1.%s Install Envoy Gateway %s(%s)%s with AI Gateway values\n" "$C_PURPLE" "$C_RESET" "$C_DIM" "$EG_VERSION" "$C_RESET"
  printf "  %s2.%s Install AI Gateway CRDs %s(%s)%s\n" "$C_PURPLE" "$C_RESET" "$C_DIM" "$AIGW_VERSION" "$C_RESET"
  printf "  %s3.%s Install AI Gateway controller %s(%s)%s\n" "$C_PURPLE" "$C_RESET" "$C_DIM" "$AIGW_VERSION" "$C_RESET"
  printf "\nTarget cluster: %s%s%s\n\n" "$C_BOLD" "$(kubectl config current-context 2>/dev/null || echo 'unknown')" "$C_RESET"

  if ! confirm "Proceed with install?"; then
    err "Aborted by user."
    exit 1
  fi
}

# ---- Install steps ----------------------------------------------------------

install_envoy_gateway() {
  info "[1/3] Installing Envoy Gateway ($EG_VERSION)"
  helm upgrade -i eg oci://docker.io/envoyproxy/gateway-helm \
    --version "$EG_VERSION" \
    --namespace envoy-gateway-system --create-namespace \
    -f "$EG_VALUES_URL"
  kubectl wait --timeout=2m -n envoy-gateway-system \
    deployment/envoy-gateway --for=condition=Available
  ok "Envoy Gateway ready"
}

install_aigw_crds() {
  info "[2/3] Installing AI Gateway CRDs ($AIGW_VERSION)"
  helm upgrade -i aieg-crd oci://docker.io/envoyproxy/ai-gateway-crds-helm \
    --version "$AIGW_VERSION" \
    --namespace envoy-ai-gateway-system --create-namespace
  ok "CRDs installed"
}

install_aigw_controller() {
  info "[3/3] Installing AI Gateway controller ($AIGW_VERSION)"
  helm upgrade -i aieg oci://docker.io/envoyproxy/ai-gateway-helm \
    --version "$AIGW_VERSION" \
    --namespace envoy-ai-gateway-system --create-namespace
  kubectl wait --timeout=2m -n envoy-ai-gateway-system \
    deployment/ai-gateway-controller --for=condition=Available
  ok "AI Gateway controller ready"
}

# ---- Main -------------------------------------------------------------------

main() {
  printf "%s%sEnvoy AI Gateway installer%s\n" "$C_BOLD" "$C_PURPLE" "$C_RESET"
  printf "%sDocs: https://aigateway.envoyproxy.io/docs/getting-started/%s\n\n" "$C_DIM" "$C_RESET"

  preflight
  resolve_versions
  summarize_and_confirm

  install_envoy_gateway
  install_aigw_crds
  install_aigw_controller

  printf "\n%s%s✓ Envoy AI Gateway is installed.%s\n" "$C_BOLD" "$C_GREEN" "$C_RESET"
  printf "Next: %shttps://aigateway.envoyproxy.io/docs/getting-started/basic-usage%s\n" "$C_PURPLE" "$C_RESET"
}

main "$@"
