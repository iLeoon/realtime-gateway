#!/bin/zsh

# Usage: ./bench.sh ./internal/transport/tcp [component name] [dumpster file]
PKG=$1
COMPONENT=$2
FILE=$3
DATE=$(date +%Y-%m-%d)

echo "🚀 Running benchmarks for $PKG..."

RAW_OUTPUT=$(go test -bench=. -benchmem "$PKG")

echo -e "\n--- Processing Results ---"

# Extract each benchmark line and format it for the table
echo "$RAW_OUTPUT" | grep "Benchmark" | while read -r line; do
    # Column mapping for Go bench output:
    # $1: Name, $2: Iterations, $3: ns/op, $5: B/op, $7: allocs/op
    NAME=$(echo "$line" | awk -F'/' '{print $2}' | awk '{print $1}' | sed 's/-[0-9]*//')
    NS_OP=$(echo "$line" | awk '{print $3}')
    B_OP=$(echo "$line" | awk '{print $5}')
    ALLOCS=$(echo "$line" | awk '{print $7}')
    
    # Calculate Ops/sec (1,000,000,000 / ns_op)
    OPS_SEC=$(awk "BEGIN {printf \"%.0fk\", 1000000000 / $NS_OP / 1000}")

    echo "| $DATE | $COMPONENT | $NAME | $OPS_SEC | $ALLOCS | $B_OP | |"
done

echo "---------------------------------------"
echo "💡 Copy the rows above into the Table section of $FILE"

# Automatically append to the "Dumpster" at the bottom of the file
echo -e "\n<details>" >> $FILE
echo "<summary><b>Detailed Log: $DATE | $PKG</b></summary>\n" >> $FILE
echo "COMMAND: \`go test -bench=. -benchmem $PKG\`\n" >> $FILE
echo '```text' >> $FILE
echo "$RAW_OUTPUT" >> $FILE
echo '```' >> $FILE
echo "</details>" >> $FILE

echo "✅ Raw data appended to the 'Dumpster' in $FILE"
