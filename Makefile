.PHONY: build install uninstall clean bootstrap test

JFVM_BIN := jfvm
SHIM_BIN := jf
SHIM_DIR := $(HOME)/.jfvm/shim

build:
	@echo "🔧 Building jfvm CLI..."
	go build -o $(JFVM_BIN) .
	@echo "🔧 Building jf shim..."
	cd shim && go build -o $(SHIM_BIN) .

install: build
	@echo "📂 Creating shim directory: $(SHIM_DIR)"
	mkdir -p $(SHIM_DIR)
	@echo "📥 Installing binaries to $(SHIM_DIR)"
	cp $(JFVM_BIN) $(SHIM_DIR)/
	cp shim/$(SHIM_BIN) $(SHIM_DIR)/
	@echo "✅ Binaries installed."

bootstrap: install
	@echo "🔁 Checking shell config for PATH..."
	@grep -q '.jfvm/shim' ~/.bashrc 2>/dev/null || echo 'export PATH="$$HOME/.jfvm/shim:$$PATH"' >> ~/.bashrc
	@grep -q '.jfvm/shim' ~/.zshrc 2>/dev/null || echo 'export PATH="$$HOME/.jfvm/shim:$$PATH"' >> ~/.zshrc
	@grep -q '.jfvm/shim' ~/.profile 2>/dev/null || echo 'export PATH="$$HOME/.jfvm/shim:$$PATH"' >> ~/.profile
	@echo "✅ PATH updated in shell config. Run 'source ~/.bashrc' or 'source ~/.zshrc' to apply."

test: build
	@echo "🧪 Running basic functionality tests..."
	@./$(JFVM_BIN) --help > /dev/null && echo "✅ jfvm help works"
	@./$(JFVM_BIN) list > /dev/null && echo "✅ jfvm list works"
	@./$(JFVM_BIN) history > /dev/null && echo "✅ jfvm history works"
	@echo "✅ All basic tests passed!"

uninstall:
	@echo "🗑️ Removing installed binaries..."
	rm -f $(SHIM_DIR)/$(JFVM_BIN) $(SHIM_DIR)/$(SHIM_BIN)
	@echo "✅ Uninstalled."

clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f $(JFVM_BIN)
	cd shim && rm -f $(SHIM_BIN)