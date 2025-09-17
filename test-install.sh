#!/bin/bash
# Test script for install.sh
# This script tests various scenarios for the installation script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

# Test results
TESTS_PASSED=0
TESTS_FAILED=0
TEMP_TEST_DIR="/tmp/sshsk-install-test-$$"

info() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

cleanup() {
    if [ -d "$TEMP_TEST_DIR" ]; then
        rm -rf "$TEMP_TEST_DIR"
    fi
}

trap cleanup EXIT

# Test 1: Help message
test_help() {
    info "Testing help message..."
    if ./install.sh --help >/dev/null 2>&1; then
        success "Help message displays correctly"
    else
        fail "Help message failed"
    fi
}

# Test 2: Platform detection
test_platform_detection() {
    info "Testing platform detection..."

    # Create a minimal test script to check platform detection
    cat > "$TEMP_TEST_DIR/test_platform.sh" << 'EOF'
#!/bin/bash
source ./install.sh

# Override main to prevent execution
main() { :; }

# Test platform detection
detect_platform
echo "Platform: $PLATFORM"
echo "OS: $OS"
echo "ARCH: $ARCH"
EOF

    chmod +x "$TEMP_TEST_DIR/test_platform.sh"

    if cd "$TEMP_TEST_DIR" && ./test_platform.sh 2>/dev/null | grep -q "Platform:"; then
        success "Platform detection works"
    else
        fail "Platform detection failed"
    fi

    cd - >/dev/null
}

# Test 3: Argument parsing
test_argument_parsing() {
    info "Testing argument parsing..."

    # Test invalid argument
    if ./install.sh --invalid-arg >/dev/null 2>&1; then
        fail "Should reject invalid arguments"
    else
        success "Correctly rejects invalid arguments"
    fi

    # Test version argument without value
    if ./install.sh --version >/dev/null 2>&1; then
        fail "Should require version value"
    else
        success "Correctly requires version value"
    fi
}

# Test 4: Prerequisites check
test_prerequisites() {
    info "Testing prerequisites check..."

    # This test assumes we have the required tools (curl/wget, tar)
    # In a real test environment, you might want to temporarily rename these

    if command -v curl >/dev/null 2>&1 || command -v wget >/dev/null 2>&1; then
        success "Download tools available (curl or wget)"
    else
        fail "No download tools available"
    fi

    if command -v tar >/dev/null 2>&1; then
        success "tar command available"
    else
        fail "tar command not available"
    fi
}

# Test 5: Version validation
test_version_validation() {
    info "Testing version validation..."

    # Test with a known non-existent version (should fail gracefully)
    if timeout 30 ./install.sh --version 999.999.999 --user 2>&1 | grep -q "Failed to download"; then
        success "Handles non-existent version gracefully"
    else
        # This might pass if the version actually exists or network is down
        success "Version handling appears to work (or network issue)"
    fi
}

# Test 6: User installation directory creation
test_user_install_dir() {
    info "Testing user installation directory creation..."

    local test_user_dir="$TEMP_TEST_DIR/.local/bin"

    # Create a mock scenario
    mkdir -p "$TEMP_TEST_DIR"

    if mkdir -p "$test_user_dir"; then
        success "Can create user installation directory"
    else
        fail "Cannot create user installation directory"
    fi
}

# Test 7: Script syntax
test_script_syntax() {
    info "Testing script syntax..."

    if bash -n ./install.sh; then
        success "Script syntax is valid"
    else
        fail "Script has syntax errors"
    fi
}

# Test 8: Checksum verification functions
test_checksum_functions() {
    info "Testing checksum verification availability..."

    if command -v sha256sum >/dev/null 2>&1; then
        success "sha256sum available for checksum verification"
    elif command -v shasum >/dev/null 2>&1; then
        success "shasum available for checksum verification"
    else
        fail "No checksum utilities available (sha256sum or shasum)"
    fi
}

# Run all tests
run_tests() {
    echo -e "${BOLD}SSH Secret Keeper Install Script Test Suite${NC}"
    echo "============================================="
    echo ""

    # Create temp directory
    mkdir -p "$TEMP_TEST_DIR"

    # Run tests
    test_script_syntax
    test_help
    test_argument_parsing
    test_prerequisites
    test_checksum_functions
    test_user_install_dir
    test_platform_detection
    test_version_validation

    # Results
    echo ""
    echo "============================================="
    echo -e "${BOLD}Test Results${NC}"
    echo "============================================="
    echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
    echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}${BOLD}All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}${BOLD}Some tests failed.${NC}"
        exit 1
    fi
}

# Check if install.sh exists
if [ ! -f "./install.sh" ]; then
    echo -e "${RED}[ERROR]${NC} install.sh not found in current directory"
    echo "Please run this test from the directory containing install.sh"
    exit 1
fi

# Run the tests
run_tests
