# OpenAI-compatible API server demo. Boot the server, then hit it with
# a streaming chat completion piped through `streamd` so the SSE chunks
# render as live markdown.
#
# Regenerate: `vhs vhs/serve.tape` (outputs vhs/out/serve.gif)

Output vhs/out/serve.gif

Source vhs/_common.tape
Require cider
Require curl
Require streamd

Hide
Type "export PS1='\033[38;5;141m>\033[0m '"
Enter
Type "clear"
Enter
Show

Sleep 400ms

# 1. Start the server in the background.
Type "cider serve --port 8765 &"
Sleep 200ms
Enter
Sleep 1500ms

Type "clear"
Enter
Sleep 300ms

# 2. List available models. Apple's on-device FM, OpenAI-shaped.
Type "curl -s http://127.0.0.1:8765/v1/models | jq"
Sleep 200ms
Enter
Sleep 2s

Type "clear"
Enter

# 3. Stream a chat completion, render through streamd as markdown.
Type `curl -sN http://127.0.0.1:8765/v1/chat/completions \`
Enter
Type `  -H 'Content-Type: application/json' \`
Enter
Type `  -d '{"model":"apple-on-device-fm","stream":true,"messages":[`
Enter
Type `    {"role":"user","content":"Explain what makes Apple silicon fast in two short paragraphs."}`
Enter
Type `  ]}' | streamd`
Sleep 300ms
Enter
Sleep 12s

Hide
Type "kill %1 2>/dev/null; wait 2>/dev/null"
Enter
Show
