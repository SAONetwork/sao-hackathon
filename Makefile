SHELL=/usr/bin/env bash

#FFI_PATH:=extern/filecoin-ffi/
#FFI_DEPS:=.install-filcrypto
#FFI_DEPS:=$(addprefix $(FFI_PATH),$(FFI_DEPS))

#build/.filecoin-install: $(FFI_PATH)
	#$(MAKE) -C $(FFI_PATH) $(FFI_DEPS:$(FFI_PATH)%=%)
#BUILD_DEPS+=build/.filecoin-install

all: sao-ds sao-monitor sao-procnode

GOCC?=go

sao-ds:
	rm -rf sao-ds 
	$(GOCC) build -o sao-ds ./cmd/ds 

sao-monitor:
	rm -rf sao-monitor
	$(GOCC) build -o sao-monitor ./cmd/monitor

sao-procnode:
	rm -rf sao-procnode
	$(GOCC) build -o sao-procnode ./cmd/procnode

doc:
	swag init -d ./cmd/apiserver,./store,./user
