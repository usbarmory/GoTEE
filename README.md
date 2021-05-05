Introduction
============

> :warning: this is a PoC

This project demonstrates instantiation of
[TamaGo](https://github.com/f-secure-foundry/tamago) based unikernels in
privileged (PL1 - system mode) and unprivileged (PL0 - user mode), interacting
with each other through monitor/supervisor mode and custom system calls

The PoC serves as foundation to develop a
[TamaGo](https://github.com/f-secure-foundry/tamago) based Trusted Execution
Environments (TEE), with the memory safety and capabilities of Go taken to bare
metal execution within TrustZone Secure World (or equivalent technology).

While the PoC employs a Go unikernel for both privileged and unprivileged
modes, the latter can be loaded with any bare metal applet (e.g. C, Rust)
implementing the syscall API.

A compatibility layer for
[libutee](https://optee.readthedocs.io/en/latest/architecture/libraries.html#libutee)
is planned, allowing execution of [OP-TEE](https://www.op-tee.org/) compatible
applets.

![diagram](https://github.com/f-secure-foundry/GoTEE/wiki/images/diagram.jpg)

Testing
=======

```
make example_ta && make qemu
...
00:00:00 PL1 tamago/arm (go1.16.3) • TEE system/supervisor
00:00:00 PL1 loaded applet addr:0x80000000 size:1753012 entry:0x8006902c
00:00:00 PL1 starting PL0 sp:0x8fffff00 pc:0x8006902c
00:00:00 PL0 tamago/arm (go1.16.3) • TEE user applet
00:00:00 PL0 is sleeping in user mode
00:00:01 PL1 is sleeping in system mode
00:00:01 PL0 is sleeping in user mode
00:00:02 PL1 is sleeping in system mode
00:00:02 PL0 is sleeping in user mode
00:00:03 PL1 is sleeping in system mode
00:00:03 PL0 is sleeping in user mode
00:00:04 PL1 is sleeping in system mode
00:00:04 PL0 is sleeping in user mode
00:00:05 PL1 is sleeping in system mode
00:00:05 PL0 about to read PL1 memory at 0x90010000
00:00:05 stopped PL0 task sp:0x81428f3c lr:0x8009f840 pc:0x80011374 err:exception mode ABT
00:00:06 PL1 is sleeping in system mode
00:00:07 PL1 is sleeping in system mode
00:00:08 PL1 is sleeping in system mode
...
```

Debugging
=========

```
make example_ta && make qemu-gdb
```

```
arm-none-eabi-gdb -ex "target remote 127.0.0.1:1234" os
>>> add-symbol-file examples/ta
>>> b ta.go:main.main
```

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com | andrea@inversepath.com  

License
=======

Copyright (c) F-Secure Corporation

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation under version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU General Public License for more details.

See accompanying LICENSE file for full details.
