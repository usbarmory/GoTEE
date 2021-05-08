Introduction
============

> :warning: this project is at PoC stage

This project demonstrates concurrent instantiation of
[TamaGo](https://github.com/f-secure-foundry/tamago) based unikernels in
privileged and unprivileged modes, interacting with each other through
monitor/supervisor mode and custom system calls

GoTEE aims to implement a [TamaGo](https://github.com/f-secure-foundry/tamago)
based Trusted Execution Environments (TEE), bringing Go memory safety,
convenience and capabilities to bare metal execution within TrustZone Secure
World or equivalent isolation technology.

In TEE nomenclature, the privileged unikernel is commonly referred to as
trusted OS, while the unprivileged one represents a trusted applet (TA).

While the repository examples implement a Go unikernel for both the trusted OS
and TA either can be replaced with any bare metal code (e.g.  C, Rust)
implementing the syscall API.

A compatibility layer for
[libutee](https://optee.readthedocs.io/en/latest/architecture/libraries.html#libutee)
is planned, allowing execution of/as [OP-TEE](https://www.op-tee.org/)
compatible applets.

![diagram](https://github.com/f-secure-foundry/GoTEE/wiki/images/diagram.jpg)

Operation
=========

The example trusted OS/applet combination performs basic testing of concurrent
execution of two [TamaGo](https://github.com/f-secure-foundry/tamago)
unikernels at different privilege levels:

 * PL1 / privileged system+supervisor mode: trusted OS
 * PL2 / unprivileged user mode: trusted applet

The TA sleeps for 5 seconds before attempting to read privileged OS memory,
which triggers an exception handled by the supervisor which terminates the TA.

A basic [syscall](https://github.com/f-secure-foundry/GoTEE/blob/master/syscall/syscall.go)
interface is implemented for communication between the two privilege levels.

Compiling
=========

Build the [TamaGo compiler](https://github.com/f-secure-foundry/tamago-go)
(or use the [latest binary release](https://github.com/f-secure-foundry/tamago-go/releases/latest)):

```
wget https://github.com/f-secure-foundry/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Build the example trusted applet and kernel executables:

```
make example_ta && make example_os
```

Executing and debugging
=======================

Native hardware
---------------

The PoC can be executed on the [USB armory Mk II](https://github.com/f-secure-foundry/usbarmory/wiki)
by loading the compilation output `os.imx` [SDP mode](https://github.com/f-secure-foundry/usbarmory/wiki/Boot-Modes-(Mk-II)#serial-download-protocol-sdp),
(note that for now the PoC only provides serial console feedback).

Emulated hardware
-----------------

An emulated run under QEMU can be executed as follows:

```
make example_ta && make example_os && make qemu
...
00:00:00 PL1 tamago/arm (go1.16.4) • TEE system/supervisor
00:00:00 PL1 loaded applet addr:0x82000000 size:1756695 entry:0x82069080
00:00:00 PL1 will sleep until PL0 is done
00:00:00 PL1 starting PL0 sp:0x83ffff00 pc:0x82069080
00:00:00 PL0 tamago/arm (go1.16.4) • TEE user applet
00:00:00 PL0 will sleep for 5 seconds
00:00:01 PL1 says 1 missisipi
00:00:01 PL0 says 1 missisipi
...
00:00:05 PL1 says 5 missisipi
00:00:05 PL0 says 5 missisipi
00:00:05 PL0 about to read PL1 memory at 0x80010000
00:00:05        r0:80010000   r1:824220c0   r2:00000001   r3:00000000
00:00:05        r1:824220c0   r2:00000001   r3:00000000   r4:00000000
00:00:05        r5:00000000   r6:00000000   r7:00000000   r8:00000007
00:00:05        r9:00000034  r10:824000e0  r11:800fbedc  r12:00000000
00:00:05        sp:82428f30   lr:8209fac8   pc:82011374 spsr:600000d0
00:00:05 PL1 stopped PL0 task sp:0x82428f30 lr:0x8209fac8 pc:0x82011374 err:exception mode ABT
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
