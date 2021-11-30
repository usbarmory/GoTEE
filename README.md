Introduction
============

The [GoTEE](https://github.com/f-secure-foundry/GoTEE) framework implements concurrent instantiation of
[TamaGo](https://github.com/f-secure-foundry/tamago) based unikernels in
privileged and unprivileged modes, interacting with each other through monitor
mode and custom system calls.

With these capabilities GoTEE implements a [TamaGo](https://github.com/f-secure-foundry/tamago)
based Trusted Execution Environments (TEE), bringing Go memory safety,
convenience and capabilities to bare metal execution within TrustZone Secure
World or equivalent isolation technology.

GoTEE can supervise pure Go, Rust or C based freestanding Trusted Applets,
implementing the GoTEE API, as well as any operating system capable of running
in TrustZone Normal World such as Linux.

A compatibility layer for
[libutee](https://optee.readthedocs.io/en/latest/architecture/libraries.html#libutee)
is planned, allowing execution of/as [OP-TEE](https://www.op-tee.org/)
compatible applets.

<img src="https://github.com/f-secure-foundry/GoTEE/wiki/images/diagram.jpg" width="350">

Documentation
=============

The main documentation, which includes a tutorial, can be found on the
[project wiki](https://github.com/f-secure-foundry/GoTEE/wiki).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/f-secure-foundry/GoTEE).

Supported hardware
==================

The following table summarizes currently supported SoCs and boards.

| SoC           | Board                                                                                                                                                                                | SoC package                                                                   | Board package                                                                                          |
|---------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------|
| NXP i.MX6ULZ  | [USB armory Mk II](https://github.com/f-secure-foundry/usbarmory/wiki)                                                                                                               | [imx6](https://github.com/f-secure-foundry/tamago/tree/master/soc/imx6)       | [usbarmory/mark-two](https://github.com/f-secure-foundry/tamago/tree/master/board/f-secure/usbarmory)  |
| NXP i.MX6ULL  | [MCIMX6ULL-EVK](https://www.nxp.com/design/development-boards/i-mx-evaluation-and-development-boards/evaluation-kit-for-the-i-mx-6ull-and-6ulz-applications-processor:MCIMX6ULL-EVK) | [imx6](https://github.com/f-secure-foundry/tamago/tree/master/soc/imx6)       | [mx6ullevk](https://github.com/f-secure-foundry/tamago/tree/master/board/nxp/mx6ullevk)                |

Example application
===================

In TEE nomenclature, the privileged unikernel is commonly referred to as
Trusted OS, while the unprivileged one represents a Trusted Applet.

The GoTEE [example](https://github.com/f-secure-foundry/GoTEE-example)
demonstrate concurrent operation of Go unikernels acting as Trusted OS,
Trusted Applet and Main OS.

> :warning: the Main OS can be any "rich" OS (e.g. Linux), TamaGo is simply
> used for a self-contained example. The same applies to the Trusted Applet
> which can be any bare metal application capable of running in user mode and
> implementing GoTEE API, such as freestanding C or Rust programs.
>
> A Rust [example](https://github.com/f-secure-foundry/GoTEE-example/tree/master/trusted_applet_rust)
> can be used replacing `trusted_applet_go` with `trusted_applet_rust` when building.

The example trusted OS/applet combination performs basic testing of concurrent
execution of three [TamaGo](https://github.com/f-secure-foundry/tamago)
unikernels at different privilege levels:

 * Trusted OS, running in Secure World at privileged level (PL1, system mode)
 * Trusted Applet, running in Secure World at unprivileged level (PL0, user mode)
 * Main OS, running in Normal World at privileged level (PL1, system mode)

The Main OS yields back with a monitor call.

The Trusted Applet sleeps for 5 seconds before attempting to read privileged OS
memory, which triggers an exception handled by the supervisor which terminates
the Trusted Applet.

The GoTEE [syscall](https://github.com/f-secure-foundry/GoTEE/blob/master/syscall/syscall.go)
interface is implemented for communication between the Trusted OS and Trusted
Applet.

When launched on the [USB armory Mk II](https://github.com/f-secure-foundry/usbarmory/wiki),
the example application is reachable via SSH through
[Ethernet over USB](https://github.com/f-secure-foundry/usbarmory/wiki/Host-communication)
(ECM protocol, supported on Linux and macOS hosts):

```
$ ssh gotee@10.0.0.1
PL1 tamago/arm (go1.16.5) • TEE system/monitor (Secure World)

  help                                   # this help
  reboot                                 # reset the SoC/board
  stack                                  # stack trace of current goroutine
  stackall                               # stack trace of all goroutines
  md  <hex offset> <size>                # memory display (use with caution)
  mw  <hex offset> <hex value>           # memory write   (use with caution)

  gotee                                  # TrustZone example w/ TamaGo unikernels
  linux <uSD|eMMC>                       # Boot NonSecure USB armory Debian base image

  dbg                                    # show ARM debug permissions
  csl                                    # show config security levels (CSL)
  csl <periph> <slave> <hex csl>         #  set config security level  (CSL)
  sa                                     # show security access (SA)
  sa  <id> <secure|nonsecure>            #  set security access (SA)

>
```

The example can be launched with the `gotee` command which spawns the Main OS
twice to demonstrate behaviour before and after TrustZone restrictions are in
effect using real hardware peripherals.

Additionally the `linux` command can be used to spawn the
[USB armory Debian base image](https://github.com/f-secure-foundry/usbarmory-debian-base_image)
as Main OS in NonSecure World.

> :warning: only USB armory Debian base image releases >= 20211129 are
> supported for NonSecure World operation.

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
git clone https://github.com/f-secure-foundry/GoTEE-example
cd GoTEE-example && make nonsecure_os_go && make trusted_applet_go && make trusted_os
```

> :warning: replace `trusted_applet_go` with `trusted_applet_rust` for a Rust
> TA example, this requires Rust nightly and the `armv7a-none-eabi` toolchain.

All compilation outputs are written under a `bin` subdirectory.

Executing and debugging
=======================

Native hardware
---------------

The PoC can be executed on the [USB armory Mk II](https://github.com/f-secure-foundry/usbarmory/wiki)
by loading the compilation output `bin/trusted_os.imx` in
[SDP mode](https://github.com/f-secure-foundry/usbarmory/wiki/Boot-Modes-(Mk-II)#serial-download-protocol-sdp).

![gotee](https://github.com/f-secure-foundry/GoTEE/wiki/images/gotee.png)

Emulated hardware
-----------------

An emulated run of the previously compiled executables under QEMU can be
executed as follows:

```
make qemu
...
00:00:00 PL1 tamago/arm (go1.17.1) • TEE system/monitor (Secure World)
00:00:00 PL1 loaded applet addr:0x94000000 size:4093296 entry:0x9406e3a4
00:00:00 PL1 loaded kernel addr:0x80000000 size:3715053 entry:0x8006cd34
00:00:00 PL1 waiting for applet and kernel
00:00:00 PL1 starting mode:USR ns:false sp:0x96000000 pc:0x9406e3a4
00:00:00 PL1 starting mode:SYS ns:true sp:0x00000000 pc:0x8006cd34
00:00:00 PL1 tamago/arm (go1.17.1) • system/supervisor (Normal World)
00:00:00 PL1 in Normal World is about to yield back
00:00:00    r0:00000000  r1:814223f0  r2:00000001  r3:00000000
00:00:00    r4:00000000  r5:00000000  r6:00000000  r7:00000000
00:00:00    r8:00000007  r9:00000034 r10:814000f0 r11:802885a9 cpsr:600001d6 (MON)
00:00:00   r12:00000000  sp:81441f88  lr:80146ac8  pc:801444c0 spsr:600000df (SYS)
00:00:00 PL1 stopped mode:SYS ns:true sp:0x81441f88 lr:0x80146ac8 pc:0x801444c0 err:exit
00:00:00 PL0 tamago/arm (go1.17.1) • TEE user applet (Secure World)
00:00:00 PL0 obtained 16 random bytes from PL1: 431ad651c0c8fb66f929df143eed6411
00:00:00 PL0 requests echo via RPC: hello
00:00:00 PL0 received echo via RPC: hello
00:00:00 PL0 will sleep for 5 seconds
00:00:01 PL1 says 1 missisipi
00:00:01 PL0 says 1 missisipi
...
00:00:05 PL1 says 5 missisipi
00:00:05 PL0 says 5 missisipi
00:00:05 PL0 is about to read PL1 Secure World memory at 0x90010000
00:00:05    r0:90010000  r1:948220c0  r2:90010000  r3:00000000
00:00:05    r4:00000000  r5:00000000  r6:00000000  r7:00000000
00:00:05    r8:00000007  r9:00000044 r10:948000f0 r11:942cba81 cpsr:600001d7 (ABT)
00:00:05   r12:00000000  sp:9484ff04  lr:94168868  pc:9401132c spsr:600000d0 (USR)
00:00:05 PL1 stopped mode:USR ns:false sp:0x9484ff04 lr:0x94168868 pc:0x9401132c err:ABT
00:00:05 PL1 says goodbye
```

> :warning: the emulated run performs partial tests due to lack of full
> TrustZone support by QEMU.

Debugging
---------

```
make qemu-gdb
```

```
arm-none-eabi-gdb -ex "target remote 127.0.0.1:1234" bin/trusted_os.elf
>>> add-symbol-file bin/trusted_applet.elf
>>> b main.main
```

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com | andrea@inversepath.com  

Andrej Rosano  
andrej.rosano@f-secure.com | andrej@inversepath.com  

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
