PREFIX?=$(shell pwd)

.DEFAULT: all
all: pkg/pkg net/net

pkg/pkg: 
	@echo "+ $@"
	cd pkg && make

net/net:
	@echo "+ $@"
	cd net && make
