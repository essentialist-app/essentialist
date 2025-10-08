
# Makefile for essentialist and flashdown

# Go parameters
GOCMD=go

# Flatpak support
FLATPAK ?= false
ifeq ($(FLATPAK), true)
	GOTAGS += flatpak
endif

# Wayland support
WAYLAND ?= false
ifeq ($(WAYLAND), true)
	GOTAGS += wayland
endif

# Binaries
ESSENTIALIST_BIN=essentialist
FLASHDOWN_BIN=flashdown

# Install directories
PREFIX?=/usr/local
BINDIR?=$(PREFIX)/bin
DATADIR?=$(PREFIX)/share
METADIR?=$(DATADIR)/metainfo
APPDIR?=$(DATADIR)/applications
SVGICONSDIR?=$(DATADIR)/icons/hicolor/scalable/apps
PNGICONSDIR?=$(DATADIR)/icons/hicolor/512x512/apps

all: $(ESSENTIALIST_BIN) $(FLASHDOWN_BIN)

$(ESSENTIALIST_BIN):
	$(GOCMD) build -tags="$(GOTAGS)" -o $(ESSENTIALIST_BIN) ./cmd/essentialist

$(FLASHDOWN_BIN):
	$(GOCMD) build -o $(FLASHDOWN_BIN) ./cmd/flashdown

clean:
	rm -f $(ESSENTIALIST_BIN) $(FLASHDOWN_BIN)

install: all
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 $(ESSENTIALIST_BIN) $(DESTDIR)$(BINDIR)
	install -m 755 $(FLASHDOWN_BIN) $(DESTDIR)$(BINDIR)
	install -d $(DESTDIR)$(APPDIR)
	install -m 644 cmd/essentialist/flatpak/io.github.essentialist_app.essentialist.desktop $(DESTDIR)$(APPDIR)/
	sed -i 's|Exec=essentialist|Exec=$(DESTDIR)$(BINDIR)/$(ESSENTIALIST_BIN)|' $(DESTDIR)$(APPDIR)/io.github.essentialist_app.essentialist.desktop
	install -d $(DESTDIR)$(SVGICONSDIR)
	install -m 644 cmd/essentialist/flatpak/io.github.essentialist_app.essentialist.svg $(DESTDIR)$(SVGICONSDIR)/
	install -d $(DESTDIR)$(PNGICONSDIR)
	install -m 644 cmd/essentialist/flatpak/io.github.essentialist_app.essentialist.png $(DESTDIR)$(PNGICONSDIR)/
	install -d $(DESTDIR)$(METADIR)
	install -m 644 cmd/essentialist/flatpak/io.github.essentialist_app.essentialist.metainfo.xml $(DESTDIR)$(METADIR)/


flatpak:
	@echo "--> Making flatpak..."
	@sed -i '/      - type: git/,/        tag:/d' cmd/essentialist/flatpak/io.github.essentialist_app.essentialist.yml
	@sed -i '/    sources:/a\      - type: dir\n        path: ../../..' cmd/essentialist/flatpak/io.github.essentialist_app.essentialist.yml
	@(cd cmd/essentialist/flatpak && flatpak-builder --force-clean --repo=repo build-dir io.github.essentialist_app.essentialist.yml)
	@(cd cmd/essentialist/flatpak && flatpak build-bundle repo ../../../essentialist.flatpak io.github.essentialist_app.essentialist)
	@echo "--> Reverting changes to flatpak config file..."
	@git checkout -- cmd/essentialist/flatpak/io.github.essentialist_app.essentialist.yml


.PHONY: all clean install flatpak
