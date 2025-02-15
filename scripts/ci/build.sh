#!/bin/bash

# Exit immediately on failure.
set -e

# Determine the operating system.
MUTAGEN_OS_NAME="$(go env GOOS)"

# Perform a build that's appropriate for the platform.
if [[ "${MUTAGEN_OS_NAME}" == "darwin" ]]; then
    # Compute the path and password for a temporary keychain where we'll import
    # the macOS code signing certificate and private key.
    MUTAGEN_KEYCHAIN_PATH="${RUNNER_TEMP}/mutagen.keychain-db"
    MUTAGEN_KEYCHAIN_PASSWORD="$(dd if=/dev/random bs=1024 count=1 2>/dev/null | openssl dgst -sha256)"

    # Store the previous default keychain.
    PREVIOUS_DEFAULT_KEYCHAIN="$(security default-keychain | xargs)"

    # Create the temporary keychain, set it to be the default keychain, set it
    # to automatically re-lock (just in case removal fails), and unlock it.
    security create-keychain -p "${MUTAGEN_KEYCHAIN_PASSWORD}" "${MUTAGEN_KEYCHAIN_PATH}"
    security default-keychain -s "${MUTAGEN_KEYCHAIN_PATH}"
    security set-keychain-settings -lut 3600 "${MUTAGEN_KEYCHAIN_PATH}"
    security unlock-keychain -p "${MUTAGEN_KEYCHAIN_PASSWORD}" "${MUTAGEN_KEYCHAIN_PATH}"

    # Import the macOS code signing certificate and private key and allow access
    # from the codesign utility.
    MUTAGEN_CERTIFICATE_AND_KEY_PATH="${RUNNER_TEMP}/certificate_and_key.p12"
    echo -n "${MACOS_CODESIGN_CERTIFICATE_AND_KEY}" | base64 --decode --output "${MUTAGEN_CERTIFICATE_AND_KEY_PATH}"
    security import "${MUTAGEN_CERTIFICATE_AND_KEY_PATH}" -k "${MUTAGEN_KEYCHAIN_PATH}" -P "${MACOS_CODESIGN_CERTIFICATE_AND_KEY_PASSWORD}" -T "/usr/bin/codesign"
    rm "${MUTAGEN_CERTIFICATE_AND_KEY_PATH}"
    security set-key-partition-list -S apple-tool:,apple: -s -k "${MUTAGEN_KEYCHAIN_PASSWORD}" "${MUTAGEN_KEYCHAIN_PATH}" > /dev/null

    # Perform a full release build.
    go run scripts/build.go --mode=release --macos-codesign-identity="${MACOS_CODESIGN_IDENTITY}"

    # Determine the Mutagen version.
    MUTAGEN_VERSION="$(build/mutagen version)"

    # Convert the windows/386 bundle to zip format.
    tar xzf "build/release/mutagen_windows_386_v${MUTAGEN_VERSION}.tar.gz"
    zip "build/release/mutagen_windows_386_v${MUTAGEN_VERSION}.zip" mutagen.exe mutagen-agents.tar.gz
    rm mutagen.exe mutagen-agents.tar.gz

    # Convert the windows/amd64 bundle to zip format.
    tar xzf "build/release/mutagen_windows_amd64_v${MUTAGEN_VERSION}.tar.gz"
    zip "build/release/mutagen_windows_amd64_v${MUTAGEN_VERSION}.zip" mutagen.exe mutagen-agents.tar.gz
    rm mutagen.exe mutagen-agents.tar.gz

    # Convert the windows/arm bundle to zip format.
    tar xzf "build/release/mutagen_windows_arm_v${MUTAGEN_VERSION}.tar.gz"
    zip "build/release/mutagen_windows_arm_v${MUTAGEN_VERSION}.zip" mutagen.exe mutagen-agents.tar.gz
    rm mutagen.exe mutagen-agents.tar.gz

    # Convert the windows/arm64 bundle to zip format.
    tar xzf "build/release/mutagen_windows_arm64_v${MUTAGEN_VERSION}.tar.gz"
    zip "build/release/mutagen_windows_arm64_v${MUTAGEN_VERSION}.zip" mutagen.exe mutagen-agents.tar.gz
    rm mutagen.exe mutagen-agents.tar.gz

    # Reset the default keychain and remove the temporary keychain.
    security default-keychain -s "${PREVIOUS_DEFAULT_KEYCHAIN}"
    security delete-keychain "${MUTAGEN_KEYCHAIN_PATH}"
else
    go run scripts/build.go --mode=slim
fi

# Ensure that the sidecar entrypoint builds.
go build ./cmd/mutagen-sidecar

# Build test scripts to ensure that they are maintained as core packages evolve.
go build ./scripts/scan_bench
go build ./scripts/watch_demo
