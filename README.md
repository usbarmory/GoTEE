Introduction
============

> :warning: this is a PoC

This project demonstrates concurrent instantiation of
[TamaGo](https://github.com/f-secure-foundry/tamago) based unikernels in
privileged and unprivileged modes, interacting with each other through
monitor/supervisor mode and custom system calls

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

Operation
=========

The PoC performs basic testing of concurrent execution of two
[TamaGo](https://github.com/f-secure-foundry/tamago) unikernels at
different privilege levels:

 * PL1 / system mode: trusted OS
 * PL2 / user mode: trusted applet

The trusted applet sleeps for 5 seconds before attempting to read privileged
memory, which is used to test exception handling by the supervisor.

A basic [syscall](https://github.com/f-secure-foundry/GoTEE/blob/master/syscall/syscall.go)
interface is implemented for communication between the two processes.

Executing
=========

For real hardware the PoC can be compiled for the [USB armory Mk II](https://github.com/f-secure-foundry/usbarmory/wiki)
as follows:

```
make example_ta && make example_os

```

The resulting `os.imx` can be executed via
[SDP mode](https://github.com/f-secure-foundry/usbarmory/wiki/Boot-Modes-(Mk-II)#serial-download-protocol-sdp),
(note that for now the PoC only provides serial console feedback).

Emulating
=========

An emulated run under QEMU can be performed as follows:

```
make example_ta && make qemu
...
00:00:00 PL1 tamago/arm (go1.16.3) • TEE system/supervisor
00:00:00 PL1 loaded applet addr:0x80000000 size:1752915 entry:0x80068fac
00:00:00 PL1 will sleep until PL0 is done
00:00:00 PL1 starting PL0 sp:0x8fffff00 pc:0x80068fac
00:00:00 PL0 tamago/arm (go1.16.3) • TEE user applet
00:00:00 PL0 will sleep for 5 seconds
00:00:01 PL1 says 1 missisipi
00:00:01 PL0 says 1 missisipi
...
00:00:05 PL1 says 5 missisipi
00:00:05 PL0 says 5 missisipi
00:00:05 PL0 about to read PL1 memory at 0x90010000
00:00:05        r0:90010000   r1:814220c0   r2:00000001   r3:00000000
00:00:05        r1:814220c0   r2:00000001   r3:00000000   r4:00000000
00:00:05        r5:00000000   r6:00000000   r7:00000000   r8:00000007
00:00:05        r9:00000034  r10:814000e0  r11:900fbe44  r12:00000000
00:00:05        sp:81428f30   lr:8009f7c0   pc:80011374 spsr:600000d0
00:00:05 PL1 stopped PL0 task sp:0x81428f30 lr:0x8009f7c0 pc:0x80011374 err:exception mode ABT
00:00:05 PL1 says goodbye
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
