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
# Pre-build the demo. Sleep is generous so go build finishes before
# we clear; otherwise its output can leak into the recorded frames.
Sleep 8s
Type "clear"
Enter
Sleep 200ms
Show

Sleep 400ms

# `bat -pp` is the shortcut for --plain --paging=never. Keeping the
# command short matters: the VHS terminal is ~113 cols wide and a
# longer command wraps into a junk-looking second line.
# Lines 12-25 cover the CalculatorArgs struct (json/description/enum
# tags) and the start of the fm.FuncTool call.
Type "bat -pp --line-range 12:25 examples/tool/main.go"
Sleep 200ms
Enter
Sleep 6s

Type "clear"
Enter
Sleep 300ms

Type "/tmp/tool-demo"
Sleep 200ms
Enter
Sleep 6s

Hide
Type "rm -f /tmp/tool-demo"
Enter
Show
