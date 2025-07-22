.PHONY: all build run clean open png

BINARY := compose-viz
SRC := ./cmd/compose-viz
COMPOSE := complicated-compose.yml
OUT := composeGraph
DOT := $(OUT).dot
PNG := $(OUT).png

all: build

build:
	go build -o $(BINARY) $(SRC)

run: build
	./$(BINARY) -v -f=$(COMPOSE) -o=$(OUT) -format=dot

png: build
	./$(BINARY) -v -f=$(COMPOSE) -o=$(OUT) -format=png

open: png
	@case "$$(uname)" in \
		Darwin*) open $(PNG) ;; \
		Linux*) xdg-open $(PNG) ;; \
		*) echo "No open command for this OS" ;; \
	esac

clean:
	rm -f $(BINARY) $(DOT) $(PNG)
