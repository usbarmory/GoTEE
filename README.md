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
execution of three [TamaGo](https://github.com/f-secure-foundry/tamago)
unikernels at different privilege levels:

 * Trusted OS, running in Secure World at privileged level (PL1, system mode)
 * Trusted Applet, running in Secure World at unprivileged level (PL0, user mode)
 * Main OS, running in NonSecure World at privileged level (PL1, system mode)

> :warning: the main OS can be anything (e.g. Linux), TamaGo is simply used for
> a self-contained example.

The TA sleeps for 5 seconds before attempting to read privileged OS memory,
which triggers an exception handled by the supervisor which terminates the TA.

A basic [syscall](https://github.com/f-secure-foundry/GoTEE/blob/master/syscall/syscall.go)
interface is implemented for communication between the two privilege levels.

Status
======

- [x] PL0/PL1 separation
- [ ] PL0 virtual address space
- [x] PL0/PL1 base syscall API
- [x] PL0/PL1 user net/rpc API
- [ ] PL0/PL1 GoTEE crypto API
- [x] Secure World execution
- [x] Normal World execution
- [ ] Normal World isolation
- [ ] Secure/Normal World API
- [ ] TEE Client API

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
make example_ta && make example_ns && make example_os
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
make example_ta && make example_ns && make example_os && make qemu
...
00:00:00 PL1 tamago/arm (go1.16.4) • TEE system/supervisor (Secure World)
00:00:00 PL1 loaded applet addr:0x82000000 size:3897006 entry:0x8206dab8
00:00:00 PL1 loaded kernel addr:0x84000000 size:3785829 entry:0x8406c6c4
00:00:00 PL1 will sleep until applet and kernel are done
00:00:00 PL1 starting mode:USR ns:false sp:0x00000000 pc:0x8206dab8
00:00:00 PL1 starting mode:SYS ns:true  sp:0x00000000 pc:0x8406c6c4
00:00:00 PL1 tamago/arm (go1.16.4) • system/supervisor (NonSecure World)
00:00:00        r0:00000000   r1:848220c0   r2:00000001   r3:00000000
00:00:00        r1:848220c0   r2:00000001   r3:00000000   r4:00000000
00:00:00        r5:00000000   r6:00000000   r7:00000000   r8:00000007
00:00:00        r9:0000004b  r10:848000e0  r11:802bd9f0  r12:00000000
00:00:00        sp:8484ff84   lr:8414b10c   pc:841471b4 spsr:600000df
00:00:00 PL1 stopped mode:SYS ns:true sp:0x8484ff84 lr:0x8414b10c pc:0x841471b4 err:exception mode MON
00:00:00 PL0 tamago/arm (go1.16.4) • TEE user applet (Secure World)
00:00:00 PL0 obtained 16 random bytes from PL1: 0ccee42855937b096ad13f395ba6f633
00:00:00 PL0 requests echo via RPC: hello
00:00:00 PL0 received echo via RPC: hello
00:00:00 PL0 will sleep for 5 seconds
00:00:01 PL1 says 1 missisipi
00:00:01 PL0 says 1 missisipi
...
00:00:05 PL1 says 5 missisipi
00:00:05 PL0 says 5 missisipi
00:00:05 PL0 is about to read PL1 memory at 0x80010000
00:00:05        r0:80010000   r1:828220c0   r2:00000001   r3:00000000
00:00:05        r1:828220c0   r2:00000001   r3:00000000   r4:00000000
00:00:05        r5:00000000   r6:00000000   r7:00000000   r8:00000007
00:00:05        r9:00000037  r10:828000e0  r11:802bd9f0  r12:00000000
00:00:05        sp:8284df2c   lr:821593d8   pc:82011374 spsr:600000d0
00:00:05 PL1 stopped mode:USR ns:false sp:0x8284df2c lr:0x821593d8 pc:0x82011374 err:exception mode ABT
00:00:05 PL1 says goodbye
```

Debugging
=========

```
make example_ta && make example_ns && make qemu-gdb
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
