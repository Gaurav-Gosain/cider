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
Type "clear"
Enter
Show

Sleep 300ms

Type "bat -p --paging=never examples/extract/main.go 2>/dev/null || cat examples/extract/main.go"
Sleep 200ms
Enter
Sleep 5s

Type "clear"
Enter

Type "/tmp/extract-demo"
Sleep 200ms
Enter
Sleep 4s

Hide
Type "rm -f /tmp/extract-demo"
Enter
Show
