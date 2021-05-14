module github.com/f-secure-foundry/GoTEE

go 1.16

require (
	github.com/f-secure-foundry/tamago v0.0.0-20210510185947-9f90b3ab2af1
)

replace github.com/f-secure-foundry/tamago => /mnt/git/public/tamago
replace gvisor.dev/gvisor => github.com/f-secure-foundry/gvisor v0.0.0-20210201110150-c18d73317e0f
