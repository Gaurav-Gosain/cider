# Hero tape. The chat TUI is the first thing a README visitor sees:
# launch cider, type a prompt, watch the response stream into a
# markdown-rendered card.
#
# Regenerate: `vhs vhs/chat.tape` (outputs vhs/out/chat.gif)

Output vhs/out/chat.gif

Source vhs/_common.tape
Require cider

Hide
Type "export PS1='\033[38;5;141m>\033[0m '"
Enter
Type "clear"
Enter
Show

Sleep 400ms

Type "cider"
Sleep 200ms
Enter
Sleep 1500ms

Type "Write me a haiku about Apple silicon."
Sleep 400ms
Enter
Sleep 8s

Type "Now turn it into a limerick."
Sleep 400ms
Enter
Sleep 10s

Escape
Sleep 200ms
