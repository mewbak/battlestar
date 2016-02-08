Battlestar
==========

[![Build Status](https://travis-ci.org/xyproto/battlestar.svg?branch=master)](https://travis-ci.org/xyproto/battlestar)
[![Build Status](https://drone.io/github.com/xyproto/battlestar/status.png)](https://drone.io/github.com/xyproto/battlestar/latest)

* Version: 0.4
* License: MIT
* Author: Alexander Rødseth



What is Battlestar?
-------------------

* A work in progress.
* Created for fun and for the educational process.
* A programming language specifically for 64-bit and 32-bit x86 and Linux.
* Subset of assembly with an alternative syntax and with support for inline C.
* The indended purpose is for writing 4k and 64k demoscene demos.

Features
--------

* The resulting executables are tiny!
* "hello world" is only *129* bytes on 32-bit Linux (when using sstrip from elfkickers).
* It's possible to write an operating system / kernel with only one source file.
* Full support for inline C (by utilizing gcc).
* C and Battlestar code can exist in the same source file and calls can be made both ways.
* Battlestar programs compiles almost instantly.
* Programs can be run like scripts by including this line at the top: ```#!/usr/bin/bts```
* Interrupts can be called with the same syntax for both 32-bit and 64-bit x86 on Linux.
* Also supports 16-bit x86 with DosBox.
* The intermediate assembly is fully commented.

In progress
-----------
* OS X support
* Reimplementing 16-bit demoscene demos without using any inline assembly
* See TODO

Quick installation
------------------

Make sure Go, Yasm and GCC are installed.

Install on Linux:

`sudo make install-linux`

Install on OS X:

`sudo make install-osx`

For development, install soft links instead:

`sudo make devinstall`

Build all the samples:

`make samples`

Build dependencies
------------------
* go >= 1.3

Runtime dependencies
--------------------
* yasm

Optional runtime dependencies
-----------------------------
* gcc (for inline C support)
* elftools/sstrip (for even smaller binaries)
* binutils (for disassembling with objdump)
* dosbox (for running 16-bit executables) (only GCC 4.9 and up supports compiling to 16-bit with -m16)
* SDL 2 (must be compiled and installed manually if on Red Hat 6)
* tcc (for even smaller binaries, in many cases)

