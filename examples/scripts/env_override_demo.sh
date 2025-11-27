#!/bin/bash
# Script to demonstrate environment variable overrides for config loading
# This shows how to override configuration values using environment variables

echo "=== Environment Variable Override Demo ==="
echo ""

# Example 1: Override provider type
echo "1. Override Provider Type:"
echo "   export ENGINE_PROVIDER_TYPE=postgres"
export ENGINE_PROVIDER_TYPE=postgres
echo "   ✓ Provider type set to: $ENGINE_PROVIDER_TYPE"
echo ""

# Example 2: Override formatter type
echo "2. Override Formatter Type:"
echo "   export ENGINE_FORMATTER_TYPE=csv"
export ENGINE_FORMATTER_TYPE=csv
echo "   ✓ Formatter type set to: $ENGINE_FORMATTER_TYPE"
echo ""

# Example 3: Override output type
echo "3. Override Output Type:"
echo "   export ENGINE_OUTPUT_TYPE=file"
export ENGINE_OUTPUT_TYPE=file
echo "   ✓ Output type set to: $ENGINE_OUTPUT_TYPE"
echo ""

# Example 4: Override provider parameters
echo "4. Override Provider Parameters:"
echo "   export ENGINE_PROVIDER_PARAM_HOST=localhost"
echo "   export ENGINE_PROVIDER_PARAM_PORT=5432"
echo "   export ENGINE_PROVIDER_PARAM_DATABASE=mydb"
export ENGINE_PROVIDER_PARAM_HOST=localhost
export ENGINE_PROVIDER_PARAM_PORT=5432
export ENGINE_PROVIDER_PARAM_DATABASE=mydb
echo "   ✓ Provider params set"
echo ""

# Example 5: Override formatter parameters
echo "5. Override Formatter Parameters:"
echo "   export ENGINE_FORMATTER_PARAM_INDENT=4"
echo "   export ENGINE_FORMATTER_PARAM_DELIMITER=|"
export ENGINE_FORMATTER_PARAM_INDENT=4
export ENGINE_FORMATTER_PARAM_DELIMITER="|"
echo "   ✓ Formatter params set"
echo ""

# Example 6: Override output parameters
echo "6. Override Output Parameters:"
echo "   export ENGINE_OUTPUT_PARAM_PATH=/tmp/report.json"
echo "   export ENGINE_OUTPUT_PARAM_MODE=0644"
export ENGINE_OUTPUT_PARAM_PATH=/tmp/report.json
export ENGINE_OUTPUT_PARAM_MODE=0644
echo "   ✓ Output params set"
echo ""

echo "=== All Environment Variables Set ==="
echo "Now run your Go program with:"
echo "  go run examples/config_loading/main.go"
echo ""
echo "The config file values will be overridden by these environment variables."
echo ""

# Print all ENGINE_* variables
echo "Current ENGINE_* environment variables:"
env | grep ^ENGINE_ | sort