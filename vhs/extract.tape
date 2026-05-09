# Structured-extraction demo. fm.Extract turns one prompt + a typed
# struct into a fully-typed value, schema generated from struct tags.
#
# Regenerate: `vhs vhs/extract.tape` (outputs vhs/out/extract.gif)

Output vhs/out/extract.gif

Source vhs/_common.tape
Require go

Hide
Type "export PS1='\033[38;5;141m>\033[0m '"
Enter
Type "export DYLD_LIBRARY_PATH=$PWD/foundation-models-c/.build/arm64-apple-macosx/release"
Enter
Type "go build -o /tmp/extract-demo ./examples/extract"
Enter
# Pre-build the demo. Sleep is generous so go build finishes before
# we clear; otherwise its output can leak into the recorded frames.
Sleep 8s
Type "clear"
Enter
Sleep 200ms
Show

Sleep 400ms

# `bat -pp` is the shortcut for --plain --paging=never. Short command
# avoids the VHS terminal wrap that mangles longer lines.
Type "bat -pp examples/extract/main.go"
Sleep 200ms
Enter
Sleep 5s

Type "clear"
Enter
Sleep 300ms

Type "/tmp/extract-demo"
Sleep 200ms
Enter
Sleep 4s

Hide
Type "rm -f /tmp/extract-demo"
Enter
Show
