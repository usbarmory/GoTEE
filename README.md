Introduction
============

The [GoTEE](https://github.com/usbarmory/GoTEE) framework implements concurrent instantiation of
[TamaGo](https://github.com/usbarmory/tamago) based unikernels in
privileged and unprivileged modes, interacting with each other through monitor
mode and custom system calls.

With these capabilities GoTEE implements a
[TamaGo](https://github.com/usbarmory/tamago) based Trusted Execution
Environments (TEE), bringing Go memory safety, convenience and capabilities to
bare metal execution within ARM TrustZone Secure World or RISC-V Supervisor
Execution Environments.

GoTEE can supervise pure Go, Rust or C based freestanding Trusted Applets,
implementing the GoTEE API, as well as any operating system capable of running
in ARM TrustZone Normal World or RISC-V S-mode such as Linux.

<img src="https://github.com/usbarmory/GoTEE/wiki/images/diagram.jpg" width="350">

Features
========

* [Isolated execution contexts](https://github.com/usbarmory/GoTEE/wiki/Trusted-OS-and-Applet-execution) for ARM User mode, TrustZone Normal World or RISC-V Supervisor Mode

* [Opportunistic soft lockstep for fault detection](https://github.com/usbarmory/GoTEE/wiki/Examples#opportunistic-soft-lockstep)

* [API for Trusted OS implementation](https://github.com/usbarmory/GoTEE/wiki/System-Calls#gotee-system-calls) (Syscall, JSON-RPC and exception handlers)

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/GoTEE.svg)](https://pkg.go.dev/github.com/usbarmory/GoTEE)

The main documentation, which includes a tutorial, can be found on the
[project wiki](https://github.com/usbarmory/GoTEE/wiki).

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/GoTEE).

Supported hardware
==================

The following table summarizes currently supported SoCs and boards.

| SoC          | Board                                                                                                                                                                                | SoC package                                                               | Board package                                                                        |
|--------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| NXP i.MX6ULZ | [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki)                                                                                                                      | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul)  | [usbarmory/mk2](https://github.com/usbarmory/tamago/tree/master/board/usbarmory)     |
| NXP i.MX6ULL | [MCIMX6ULL-EVK](https://www.nxp.com/design/development-boards/i-mx-evaluation-and-development-boards/evaluation-kit-for-the-i-mx-6ull-and-6ulz-applications-processor:MCIMX6ULL-EVK) | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul)  | [mx6ullevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/mx6ullevk)     |
| SiFive FU540 | [QEMU sifive_u](https://www.qemu.org/docs/master/system/riscv/sifive_u.html)                                                                                                         | [fu540](https://github.com/usbarmory/tamago/tree/master/soc/sifive/fu540) | [qemu/sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u) |

Example application
===================

In TEE nomenclature, the privileged unikernel is commonly referred to as
Trusted OS, while the unprivileged one represents a Trusted Applet.

The GoTEE [example](https://github.com/usbarmory/GoTEE-example)
demonstrate concurrent operation of Go unikernels acting as Trusted OS,
Trusted Applet and Main OS.

> [!WARNING]
> The Main OS can be any "rich" OS (e.g. Linux), TamaGo is simply
> used for a self-contained example. The same applies to the Trusted Applet
> which can be any bare metal application capable of running in user mode and
> implementing GoTEE API, such as freestanding C or Rust programs.
>
> A Rust [example](https://github.com/usbarmory/GoTEE-example/tree/master/trusted_applet_rust)
> can be used replacing `trusted_applet_go` with `trusted_applet_rust` when building.

The example trusted OS/applet combination performs basic testing of concurrent
execution of three [TamaGo](https://github.com/usbarmory/tamago)
unikernels at different privilege levels:

 * Trusted OS (ARM: TZ Secure World system mode, RISC-V: M-mode)
 * Trusted Applet (ARM: TZ Secure World user mode, RISC-V: S-mode)
 * Main OS (ARM: TZ Normal World system mode, RISC-V: S-mode)

The Main OS yields back with a monitor call.

The Trusted Applet sleeps for 5 seconds before attempting to read Trusted OS
memory, which triggers an exception handled by the supervisor which terminates
the Trusted Applet.

The GoTEE [syscall](https://github.com/usbarmory/GoTEE/blob/master/syscall/syscall.go)
interface is implemented for communication between the Trusted OS and Trusted
Applet.

When launched on the [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki),
the example application is reachable via SSH through
[Ethernet over USB](https://github.com/usbarmory/usbarmory/wiki/Host-communication)
(ECM protocol, supported on Linux and macOS hosts):

```
$ ssh gotee@10.0.0.1
tamago/arm • TEE security monitor (Secure World system/monitor)

allgptr                                          # memory forensics of applet goroutines
csl                                              # show config security levels (CSL)
csl             <periph> <slave> <hex csl>       # set config security level (CSL)
dbg                                              # show ARM debug permissions
exit, quit                                       # close session
gotee                                            # TrustZone example w/ TamaGo unikernels
help                                             # this help
linux           <uSD|eMMC>                       # boot NonSecure USB armory Debian base image
lockstep        <fault %>                        # tandem applet example w/ fault injection
peek            <hex offset> <size>              # memory display (use with caution)
poke            <hex offset> <hex value>         # memory write   (use with caution)
reboot                                           # reset device
sa                                               # show security access (SA)
sa              <id> <secure|nonsecure>          # set security access (SA)
stack                                            # stack trace of current goroutine
stackall                                         # stack trace of all goroutines

>
```

The example can be launched with the `gotee` command which spawns the Main OS
twice to demonstrate behaviour before and after TrustZone restrictions are in
effect using real hardware peripherals.

Additionally the `linux` command can be used to spawn the
[USB armory Debian base image](https://github.com/usbarmory/usbarmory-debian-base_image)
as Non-secure main OS.

> [!NOTE]
> Only USB armory Debian base image releases >= 20211129 are
> supported for Non-secure operation.

![gotee](https://github.com/usbarmory/GoTEE/wiki/images/gotee.png)

The example can be also executed under QEMU emulation.

> [!NOTE]
> Emulated runs perform partial tests due to lack of full
> TrustZone/PMP support by QEMU.

```
make qemu
...
> gotee
00:00:00 tamago/arm • TEE security monitor (Secure World system/monitor)
00:00:00 SM loaded applet addr:0x9c000000 entry:0x9c072740 size:4940275
00:00:00 SM loaded kernel addr:0x80000000 entry:0x8007100c size:4577643
00:00:00 SM waiting for applet and kernel
00:00:00 SM starting mode:USR sp:0x9e000000 pc:0x9c072740 ns:false
00:00:00 SM starting mode:SYS sp:0x00000000 pc:0x8007100c ns:true
00:00:00 tamago/arm (go1.19.1) • TEE user applet
00:00:00 tamago/arm (go1.19.1) • system/supervisor (Non-secure)
00:00:00 supervisor is about to yield back
00:00:00 SM stopped mode:SYS sp:0x8146bf54 lr:0x801937a4 pc:0x80193884 ns:true err:exit
00:00:00 applet obtained 16 random bytes from monitor: b4cc4764dd30291a52545b182313003c
00:00:00 applet requests echo via RPC: hello
00:00:00 applet received echo via RPC: hello
00:00:00 applet will sleep for 5 seconds
00:00:01 applet says 1 mississippi
...
00:00:05 applet says 5 mississippi
00:00:05 applet is about to read secure memory at 0x98010000
00:00:05    r0:98010000  r1:9c8240c0  r2:98010000  r3:00000000
00:00:05    r4:00000000  r5:00000000  r6:00000000  r7:9c86bec8
00:00:05    r8:00000007  r9:0000003d r10:9c8020f0 r11:9c342f41 cpsr:600001d7 (ABT)
00:00:05   r12:00000061  sp:9c86bf08  lr:9c1b1be8  pc:9c011330 spsr:600001d0 (USR)
00:00:05 SM stopped mode:USR sp:0x9c86bf08 lr:0x9c1b1be8 pc:0x9c011330 ns:false err:ABT
```

Building the compiler
=====================

Build the [TamaGo compiler](https://github.com/usbarmory/tamago-go)
(or use the [latest binary release](https://github.com/usbarmory/tamago-go/releases/latest)):

```
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Building and executing on ARM targets
=====================================

Build the example trusted applet and kernel executables as follows:

```
git clone https://github.com/usbarmory/GoTEE-example
cd GoTEE-example && export TARGET=usbarmory && make nonsecure_os_go && make trusted_applet_go && make trusted_os
```

> [!NOTE]
> Replace `trusted_applet_go` with `trusted_applet_rust` for a Rust
> TA example, this requires Rust nightly and the `armv7a-none-eabi` toolchain.

Final executables are created in the `bin` subdirectory,
`trusted_os_usbarmory.imx` should be used for native execution.

The following targets are available:

| `TARGET`    | Board            | Executing and debugging                                                                                  |
|-------------|------------------|----------------------------------------------------------------------------------------------------------|
| `usbarmory` | USB armory Mk II | [usbarmory](https://github.com/usbarmory/tamago/tree/master/board/usbarmory#executing-and-debugging)     |

The targets support native (see relevant documentation links in the table above)
as well as emulated execution (e.g. `make qemu`).

Building and executing on RISC-V targets
========================================

Build the example trusted applet and kernel executables as follows:

```
git clone https://github.com/usbarmory/GoTEE-example
cd GoTEE-example && export TARGET=sifive_u && make nonsecure_os_go && make trusted_applet_go && make trusted_os
```

> [!NOTE]
> Replace `trusted_applet_go` with `trusted_applet_rust` for a Rust
> TA example, this requires Rust nightly and the `riscv64gc-unknown-none-elf`
> toolchain.

Final executables are created in the `bin` subdirectory.

Available targets:

| `TARGET`    | Board            | Executing and debugging                                                                                  |
|-------------|------------------|----------------------------------------------------------------------------------------------------------|
| `sifive_u`  | QEMU sifive_u    | [sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u#executing-and-debugging)  |

The target has only been tested with emulated execution (e.g. `make qemu`)

Applications using GoTEE
========================

* [ArmoredWitness](https://github.com/transparency-dev/armored-witness) - cross-ecosystem witness network

Authors
=======

Andrea Barisani  
andrea@inversepath.com  

Andrej Rosano  
andrej@inversepath.com  

License
=======

GoTEE | https://github.com/usbarmory/GoTEE  
Copyright (c) WithSecure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/GoTEE/blob/master/LICENSE) file.
