.PHONY: all build run clean open png

BINARY          := compose-viz
SRC             := ./cmd/compose-viz
COMPOSE         ?= docker-compose-example.yml
OUT             ?= composeGraph
DOT             := $(OUT).dot
PNG             := $(OUT).png
RENDER_VOLUMES  ?= false

all: build

build:
	go build -o $(BINARY) $(SRC)

run: build
	./$(BINARY) -v -f=$(COMPOSE) -o=$(DOT) -format=dot -render-volumes=$(RENDER_VOLUMES)

png: build
	./$(BINARY) -v -f=$(COMPOSE) -o=$(PNG) -format=png -render-volumes=$(RENDER_VOLUMES)

open: png
	@case "$$(uname)" in \
		Darwin*) open $(PNG) ;; \
		Linux*) xdg-open $(PNG) ;; \
		*) echo "No open command for this OS" ;; \
	esac

clean:
	rm -f $(BINARY) $(DOT) $(PNG)
