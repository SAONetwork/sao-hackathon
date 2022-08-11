SHELL=/usr/bin/env bash

FFI_PATH:=extern/filecoin-ffi/
FFI_DEPS:=.install-filcrypto
FFI_DEPS:=$(addprefix $(FFI_PATH),$(FFI_DEPS))

build/.filecoin-install: $(FFI_PATH)
	$(MAKE) -C $(FFI_PATH) $(FFI_DEPS:$(FFI_PATH)%=%)
	@touch $@
BUILD_DEPS+=build/.filecoin-install

BINS:=
CLEAN:=
CLEAN+=build/.filecoin-install

all: sao-ds sao-monitor sao-procnode
.PHONY: all

GOCC?=go

sao-ds: $(BUILD_DEPS)
	rm -rf sao-ds
	$(GOCC) build -ldflags="-extldflags=-Wl,--allow-multiple-definition" -o sao-ds ./cmd/ds
.PHONY: sao-ds
BINS+=sao-ds

sao-monitor: $(BUILD_DEPS)
	rm -rf sao-monitor
	$(GOCC) build -o sao-monitor ./cmd/monitor
.PHONY: sao-monitor
BINS+=sao-monitor

sao-procnode: $(BUILD_DEPS)
	rm -rf sao-procnode
	$(GOCC) build -o sao-procnode ./cmd/procnode
.PHONY: sao-procnode
BINS+=sao-procnode

clean:
	rm -rf $(CLEAN) $(BINS)
	-$(MAKE) -C $(FFI_PATH) clean
.PHONY: clean
