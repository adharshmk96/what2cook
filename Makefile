.PHONY: build ui sync-ui api clean run

UI_DIR := what2cook-ui
API_DIR := what2cook-api
DIST_SRC := $(UI_DIR)/dist
DIST_DST := $(API_DIR)/web/dist
BINARY := $(API_DIR)/what2cook

# Build UI, sync into embed path, compile single binary.
build: sync-ui api

ui:
	cd $(UI_DIR) && bun run build

# Copy Vite output into the Go embed directory (web/dist).
sync-ui: ui
	rm -rf $(DIST_DST)
	mkdir -p $(DIST_DST)
	cp -R $(DIST_SRC)/. $(DIST_DST)/
	# Keep embed valid when dist is wiped in a clean tree.
	touch $(DIST_DST)/.gitkeep

api:
	cd $(API_DIR) && go build -o what2cook .

run: build
	cd $(API_DIR) && ./what2cook serve

clean:
	rm -f $(BINARY)
	rm -rf $(UI_DIR)/dist
	rm -rf $(DIST_DST)
	mkdir -p $(DIST_DST)
	touch $(DIST_DST)/.gitkeep
