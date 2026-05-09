# Tool-calling demo. fm.FuncTool[T] takes a typed handler and the model
# fills in the arguments according to the auto-generated schema.
#
# Regenerate: `vhs vhs/tool.tape` (outputs vhs/out/tool.gif)

Output vhs/out/tool.gif

Source vhs/_common.tape
Require go

Hide
Type "export PS1='\033[38;5;141m>\033[0m '"
Enter
Type "export DYLD_LIBRARY_PATH=$PWD/foundation-models-c/.build/arm64-apple-macosx/release"
Enter
Type "go build -o /tmp/tool-demo ./examples/tool"
Enter
Type "clear"
Enter
Show

Sleep 300ms

Type "bat -p --paging=never examples/tool/main.go 2>/dev/null || cat examples/tool/main.go"
Sleep 200ms
Enter
Sleep 6s

Type "clear"
Enter

Type "/tmp/tool-demo"
Sleep 200ms
Enter
Sleep 6s

Hide
Type "rm -f /tmp/tool-demo"
Enter
Show
