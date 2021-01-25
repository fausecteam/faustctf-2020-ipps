SERVICE := ipps
DESTDIR ?= dist_root
SERVICEDIR ?= /srv/$(SERVICE)
SRCS=$(shell find . -name '*.go')
.PHONY: build install

build: $(SRCS)
	go build './cmd/$(SERVICE)'

install: build
	mkdir -p $(DESTDIR)$(SERVICEDIR)
	mkdir -p $(DESTDIR)$(SERVICEDIR)/cmd
	cp -r cmd/ipps $(DESTDIR)$(SERVICEDIR)/cmd/
	cp -r README.md configs internal pkg web $(DESTDIR)$(SERVICEDIR)/
	cp configs/defaults.toml $(DESTDIR)$(SERVICEDIR)/config.toml
	cp README.md $(DESTDIR)$(SERVICEDIR)/
	cp $(SERVICE) $(DESTDIR)$(SERVICEDIR)/
	mkdir -p $(DESTDIR)/etc/systemd/system
	cp init/systemd/ipps.service $(DESTDIR)/etc/systemd/system/
	cp init/systemd/ipps-setup.service $(DESTDIR)/etc/systemd/system/
	cp init/systemd/system-ipps.slice $(DESTDIR)/etc/systemd/system/
	cd $(DESTDIR)$(SERVICEDIR) && go mod init gitlab.cs.fau.de/faust/faustctf-2020/ipps
	openssl genrsa -out $(DESTDIR)$(SERVICEDIR)/privkey.pem 2048
	openssl rsa -in $(DESTDIR)$(SERVICEDIR)/privkey.pem -outform PEM -pubout -out $(DESTDIR)$(SERVICEDIR)/pubkey.pem
