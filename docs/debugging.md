# 🛠️ Debugging

Using a debugger allows you to trace the execution flow of an application in real-time from the command line. It also enables you to pause execution at specific points to analyze the system state in detail.

## Getting Started

To run the debugger, use the `start` command with the `--debug` flag. This command will activate the debugger, allowing you to control the application's execution in detail via the command-line interface.

```sh
./dist/uniflow start --debug
```

## Commands

While the debugger is running, various commands can be used to efficiently perform debugging tasks. Below is a list of commonly used commands during the debugging process and how to utilize them.

### Quit

Use this command to exit the debugging session and return to the command line.

```sh
(debug) quit
```

This command can also be executed using `q`.

### Break

You can set breakpoints to stop code execution at specific points. This allows you to inspect the system state during execution and diagnose issues.

```sh
(debug) break                 # Set a breakpoint on all symbols
(debug) break <symbol>        # Set a breakpoint on a specific symbol
(debug) break <symbol> <port> # Set a breakpoint on a specific port of a symbol
```

This command can also be shortened to `b`.

### Continue

Use this command to resume execution from a breakpoint. The program will continue running until it reaches the next breakpoint.

```sh
(debug) continue
```

This command can also be executed using `c`.

### Delete

This command deletes set breakpoints. Each breakpoint has a unique ID, which you can use to delete a specific breakpoint.

```sh
(debug) delete              # Delete the current breakpoint
(debug) delete <breakpoint> # Delete a specific breakpoint
```

This command can also be shortened to `d`.

### Breakpoints

This command lists all currently set breakpoints. It allows you to check the location and status of each breakpoint.

```sh
(debug) breakpoints
```

This command can also be executed using `bps`.

### Breakpoint

To view detailed information about a specific breakpoint, use this command. You can query the status of a breakpoint using its ID.

```sh
(debug) breakpoint              # View details of the current breakpoint
(debug) breakpoint <breakpoint> # View details of a specific breakpoint
```

This command can also be shortened to `bp`.

### Symbols

This command lists all currently available symbols. Symbols represent nodes running at runtime, and this command allows you to view the entire list of symbols.

```sh
(debug) symbols
```

This command can also be executed using `sbs`.

### Symbol

To view detailed information about a specific symbol, use this command. By entering the symbol's ID or name, you can check its status and related information.

```sh
(debug) symbol              # View details of the current symbol
(debug) symbol <symbol>     # View details of a specific symbol
```

This command can also be shortened to `sb`.

### Processes

This command lists all currently running processes. Use it to check the status of processes operating within the system.

```sh
(debug) processes
```

This command can also be executed using `procs`.

### Process

To view detailed information about a specific process, use this command. You can check the status and related information of a process by using its ID.

```sh
(debug) process              # View details of the current process
(debug) process <process>    # View details of a specific process
```

This command can also be shortened to `proc`.

### Frames

This command lists all currently active frames. Frames represent a specific execution state of the code, and this command allows you to view the status of the frames being processed.

```sh
(debug) frames              # View frame information of the current process
(debug) frames <process>    # View frame information of a specific process
```

This command can also be executed using `frms`.

### Frame

To view detailed information about the current frame, use this command. It allows you to closely inspect the state of the frame that is currently executing.

```sh
(debug) frame
```

This command can also be shortened to `frm`.
