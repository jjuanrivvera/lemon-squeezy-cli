# lsqueezy demo — rendered with VHS (https://github.com/charmbracelet/vhs).
# No account, signup, keyring or external network: only invented local API data.
# Regenerate from the repository root:
#   go build -o bin/lsqueezy . && go build -o bin/demoserver ./tools/demoserver && vhs .demo/demo.tape
# The shell exits on a failed command rather than recording an error as a result.

Output assets/demo.gif

Set Shell bash
Set FontSize 15
Set Width 1200
Set Height 760
Set Padding 18
Set Theme "Dracula"
Set TypingSpeed 28ms

Hide
Type "set -e -o pipefail"
Enter
Type "export PATH=$PWD/bin:$PATH XDG_CONFIG_HOME=$(mktemp -d)"
Enter
Type "export LEMONSQUEEZY_API_KEY=demo-api-key"
Enter
Type `demoserver 127.0.0.1:8643 >"$XDG_CONFIG_HOME/demoserver.log" 2>&1 & demo_server_pid=$!`
Enter
Type `trap 'kill "$demo_server_pid" 2>/dev/null || true; rm -rf "$XDG_CONFIG_HOME"' EXIT`
Enter
Type `sleep 1; kill -0 "$demo_server_pid"`
Enter
Type `alias lsqueezy='env -i PATH="$PATH" XDG_CONFIG_HOME="$XDG_CONFIG_HOME" LEMONSQUEEZY_API_KEY="$LEMONSQUEEZY_API_KEY" lsqueezy --account demo --base-url http://127.0.0.1:8643/v1'`
Enter
Type "clear"
Enter
Show

Sleep 500ms

Type "lsqueezy products list"
Sleep 250ms
Enter
Sleep 2400ms

Type "lsqueezy orders list"
Sleep 250ms
Enter
Sleep 2400ms

Type "clear"
Enter
Sleep 300ms

Type "lsqueezy orders list -o json --jq '.[] | {order: .order_number, total: .total_formatted}'"
Sleep 250ms
Enter
Sleep 2400ms

Type "lsqueezy discounts create --set name='Demo launch' --set code=DEMO20 --set amount=20 --set amount_type=percent --rel store=stores:1 -o json"
Sleep 250ms
Enter
Sleep 2400ms

Type "lsqueezy discounts list"
Sleep 250ms
Enter
Sleep 2400ms
