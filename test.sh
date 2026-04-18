#!/bin/sh

set -e

REGISTRY="${REGISTRY:-reg.g5d.dev}"

total=0
passed=0
failed=0

echo "=========================================="
echo "Running all tests with registry: $REGISTRY"
echo "=========================================="
echo ""

for testfile in */test.sh; do
  [ -f "$testfile" ] || continue

  dir=$(dirname "$testfile")
  image_name="$REGISTRY/$dir"
  total=$((total + 1))

  echo -n "[$total] Testing $dir... "

  if "$testfile" "$image_name"; then
    echo "✓"
    passed=$((passed + 1))
  else
    echo "✗"
    failed=$((failed + 1))
  fi
done

echo "=========================================="
echo "Test Summary:"
echo "  Total:   $total"
echo "  Passed:  $passed"
echo "  Failed:  $failed"
echo "=========================================="

if [ $failed -eq 0 ]; then
  echo "All tests passed!"
  exit 0
else
  echo "$failed test(s) failed"
  exit 1
fi
