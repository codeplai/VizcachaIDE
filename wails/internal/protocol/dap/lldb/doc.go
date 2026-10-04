// Package lldb is the dap.Flavor of lldb-dap, the LLVM debug adapter, and the reverse handler that
// runs the debugged program in a pseudoterminal. C++ uses it today and Rust will reuse it: what
// differs between languages (the program, the folders that count as user code) is in Options.
package lldb
