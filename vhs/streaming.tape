# Library streaming demo. Run the streaming example so you can see the
# Foundation Model emitting tokens chunk-by-chunk through the Go API.
#
# Regenerate: `vhs vhs/streaming.tape` (outputs vhs/out/streaming.gif)

Output vhs/out/streaming.gif

Source vhs/_common.tape
Require go

Hide
Type "export PS1='\033[38;5;141m>\033[0m '"
Enter
Type "export DYLD_LIBRARY_PATH=$PWD/foundation-models-c/.build/arm64-apple-macosx/release"
Enter
Type "clear"
Enter
Show

Sleep 300ms

Type "go run ./examples/streaming/"
Sleep 200ms
Enter
Sleep 14s
