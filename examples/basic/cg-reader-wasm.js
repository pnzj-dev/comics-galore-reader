// cg-reader-wasm.js — JavaScript interop for CG Reader WASM
// Handles WASM initialization, DOM updates, and user interactions.

(function () {
    'use strict';

    const SHELL_ID = 'cg-reader-shell';
    const ELEMENTS = {
        shell:       document.getElementById(SHELL_ID),
        title:       document.getElementById('cg-reader-title'),
        pageInd:     document.getElementById('cg-reader-page-indicator'),
        pageImg:     document.getElementById('cg-reader-page'),
        dropZone:    document.getElementById('drop-zone'),
        fileInput:   document.getElementById('file-input'),
        browseLink:  document.getElementById('browse-link'),
        btnPrev:     document.getElementById('cg-btn-prev'),
        btnNext:     document.getElementById('cg-btn-next'),
        btnGo:       document.getElementById('cg-btn-go'),
        pageInput:   document.getElementById('cg-page-input'),
        pageTotal:   document.getElementById('cg-page-total'),
        errorBox:    document.getElementById('cg-reader-error'),
    };

    let readerID = null;
    let zoomLevel = 100;
    let comicURL = '';
    let passwordURL = '';
    let goReady = false;

    // ---- Read configuration from the shell's data attributes ----

    function readConfig() {
        const shell = ELEMENTS.shell;
        if (!shell) return;
        comicURL = shell.getAttribute('data-comic-url') || '';
        passwordURL = shell.getAttribute('data-password-url') || '';
    }

    // ---- Error display ----


    function showSpinner(text) {
        var s = document.getElementById('cg-spinner');
        var t = document.getElementById('cg-spinner-text');
        if (s) s.style.display = 'flex';
        if (t) t.textContent = text || 'Loading...';
    }
    function hideSpinner() {
        var s = document.getElementById('cg-spinner');
        if (s) s.style.display = 'none';
    }

    function showError(msg) {
        if (!ELEMENTS.errorBox) return;
        ELEMENTS.errorBox.textContent = msg;
        ELEMENTS.errorBox.style.display = 'block';
        console.error('[cg-reader]', msg);
    }

    function hideError() {
        if (!ELEMENTS.errorBox) return;
        ELEMENTS.errorBox.style.display = 'none';
    }

    // ---- UI Updates ----

    function updateUI() {
        if (!goReady || readerID === null) return;

        try {
            const pageCount = cgreader.pageCount(readerID);
            const current = cgreader.currentPage(readerID);
            const title = cgreader.title(readerID);

            if (ELEMENTS.title) {
                ELEMENTS.title.textContent = title || 'CG Reader';
            }
            if (ELEMENTS.pageInd) {
                ELEMENTS.pageInd.textContent = (current + 1) + ' / ' + pageCount;
            }
            if (ELEMENTS.pageInput) {
                ELEMENTS.pageInput.value = current + 1;
                ELEMENTS.pageInput.max = pageCount;
            }
            if (ELEMENTS.pageTotal) {
                ELEMENTS.pageTotal.textContent = pageCount;
            }

            // Enable/disable buttons.
            const hasPages = pageCount > 0;
            [ELEMENTS.btnPrev, ELEMENTS.btnNext, ELEMENTS.btnGo, ELEMENTS.pageInput].forEach(function (el) {
                if (el) el.disabled = !hasPages;
            });
            if (ELEMENTS.btnPrev) ELEMENTS.btnPrev.disabled = !hasPages || current <= 0;
            if (ELEMENTS.btnNext) ELEMENTS.btnNext.disabled = !hasPages || current >= pageCount - 1;
        } catch (e) {
            showError('UI update error: ' + e.message);
        }
    }


    function fitShellToImage() {
        var img = ELEMENTS.pageImg;
        if (!img || !img.naturalWidth || !img.naturalHeight) return;
        var naturalW = img.naturalWidth;
        var naturalH = img.naturalHeight;
        var shell = ELEMENTS.shell;
        var maxW = window.innerWidth * 0.95;
        var maxH = window.innerHeight * 0.85;
        var ratio = naturalW / naturalH;
        var w = Math.min(naturalW, maxW);
        var h = w / ratio;
        if (h > maxH) { h = maxH; w = h * ratio; }
        shell.style.maxWidth = Math.floor(w + 48) + 'px';
    }

    function renderPage() {
        hideSpinner();
        if (!goReady || readerID === null) return;

        try {
            const data = cgreader.getCurrentPage(readerID);
            if (data instanceof Uint8Array && data.length > 0) {
                const blob = new Blob([data], { type: 'image/jpeg' });
                const url = URL.createObjectURL(blob);

                // Revoke old URL to avoid memory leaks.
                const oldUrl = ELEMENTS.pageImg.src;
                if (oldUrl && oldUrl.startsWith('blob:')) {
                    URL.revokeObjectURL(oldUrl);
                }

                ELEMENTS.pageImg.src = url;
                ELEMENTS.pageImg.onload = function() {
                    updateZoom();
                    fitShellToImage();
                };
                ELEMENTS.pageImg.style.display = 'block';
                if (ELEMENTS.dropZone) {
                    ELEMENTS.dropZone.style.display = 'none';
                }
            }
        } catch (e) {
            showError('Render error: ' + e.message);
        }
    }

    // ---- Navigation ----

    function goToPage() {
        if (!goReady || readerID === null) return;
        hideError();

        const page = parseInt(ELEMENTS.pageInput.value, 10) - 1; // convert to 0-based
        if (isNaN(page)) return;

        try {
            const err = cgreader.setCurrentPage(readerID, page);
            if (err) {
                showError(err);
                return;
            }
            updateUI();
            renderPage();
        } catch (e) {
            showError(e.message);
        }
    }

    function nextPage() {
        if (!goReady || readerID === null) return;
        hideError();

        try {
            const advanced = cgreader.next(readerID);
            if (advanced) {
                updateUI();
                renderPage();
            }
        } catch (e) {
            showError(e.message);
        }
    }

    function prevPage() {
        if (!goReady || readerID === null) return;
        hideError();

        try {
            const advanced = cgreader.prev(readerID);
            if (advanced) {
                updateUI();
                renderPage();
            }
        } catch (e) {
            showError(e.message);
        }
    }

    // ---- Archive Loading ----

    function openArchive(data, filename) {
        showSpinner("Opening archive...");
        if (!goReady) {
            showError('WASM runtime not ready yet.');
            return;
        }

        hideError();

        // Close any previous reader.
        if (readerID !== null) {
            try { cgreader.close(readerID); } catch (e) { /* ignore */ }
            readerID = null;
        }

        // Create new reader with password URL from config.
        readerID = cgreader.new(passwordURL, true);

        // Convert ArrayBuffer to Uint8Array for the Go side.
        const uint8 = new Uint8Array(data);

        cgreader.openArchive(readerID, uint8, filename)
            .then(function () {
                updateUI();
                renderPage();
            })
            .catch(function (err) {
                showError('Failed to open archive: ' + err);
                readerID = null;
            });
    }

    function loadFromURL(url, filename) {
        if (!url) return;
        showSpinner('Loading comic...');
        // Temporarily show loading state.
        ELEMENTS.errorBox.style.display = 'block';
        ELEMENTS.errorBox.textContent = 'Loading comic...';
        ELEMENTS.errorBox.style.background = '#1a2a4a';
        ELEMENTS.errorBox.style.color = '#6ba3ff';

        fetch(url)
            .then(function (response) {
                if (!response.ok) throw new Error('HTTP ' + response.status);
                return response.arrayBuffer();
            })
            .then(function (buffer) {
                hideError();
                ELEMENTS.errorBox.style.background = '#4a1111';
                ELEMENTS.errorBox.style.color = '#ff6b6b';
                var name = filename || url.split('/').pop() || 'comic.cbz';
                openArchive(buffer, name);
            })
            .catch(function (err) {
                showError('Failed to fetch comic: ' + err.message);
            });
    }

    function handleFile(file) {
        if (!file) return;
        hideError();

        var reader = new FileReader();
        reader.onload = function (e) {
            openArchive(e.target.result, file.name);
        };
        reader.onerror = function () {
            showError('Failed to read file.');
        };
        reader.readAsArrayBuffer(file);
    }

    // ---- File Drop / Browse ----

    function setupFileHandling() {
        // Click to browse.
        if (ELEMENTS.browseLink && ELEMENTS.fileInput) {
            ELEMENTS.browseLink.addEventListener('click', function () {
                ELEMENTS.fileInput.click();
            });
            ELEMENTS.fileInput.addEventListener('change', function () {
                if (this.files && this.files[0]) {
                    handleFile(this.files[0]);
                }
            });
        }

        // Click on drop zone.
        if (ELEMENTS.dropZone) {
            ELEMENTS.dropZone.addEventListener('click', function () {
                if (ELEMENTS.fileInput) ELEMENTS.fileInput.click();
            });
        }

        // Drag and drop on the whole document.
        document.addEventListener('dragover', function (e) {
            e.preventDefault();
            e.stopPropagation();
        });
        document.addEventListener('drop', function (e) {
            e.preventDefault();
            e.stopPropagation();
            if (e.dataTransfer.files && e.dataTransfer.files[0]) {
                handleFile(e.dataTransfer.files[0]);
            }
        });

        // Keyboard shortcuts.
        document.addEventListener('keydown', function (e) {
            if (readerID === null) return;
            switch (e.key) {
                case 'ArrowLeft':
                    e.preventDefault();
                    prevPage();
                    break;
                case 'ArrowRight':
                    e.preventDefault();
                    nextPage();
                    break;
                case 'Home':
                    e.preventDefault();
                    if (ELEMENTS.pageInput) {
                        ELEMENTS.pageInput.value = '1';
                        goToPage();
                    }
                    break;
                case 'End':
                    e.preventDefault();
                    const total = cgreader.pageCount(readerID);
                    if (ELEMENTS.pageInput && total > 0) {
                        ELEMENTS.pageInput.value = String(total);
                        goToPage();
                    }
                    break;
            }
        });
    }

    // ---- Button Events ----


    function updateZoom() {
        var img = ELEMENTS.pageImg;
        if (img) {
            img.style.transform = 'scale(' + (zoomLevel / 100) + ')';
            img.style.transformOrigin = 'top left';
            img.style.maxWidth = 'none';
            img.style.maxHeight = 'none';
        }
        var label = document.getElementById('cg-zoom-level');
        if (label) label.textContent = zoomLevel + '%';
    }
    function zoomIn() { zoomLevel = Math.min(300, zoomLevel + 25); updateZoom(); }
    function zoomOut() { zoomLevel = Math.max(25, zoomLevel - 25); updateZoom(); }
    function zoomReset() { zoomLevel = 100; updateZoom(); }

    function setupButtons() {
        if (ELEMENTS.btnPrev) ELEMENTS.btnPrev.addEventListener('click', prevPage);
        if (ELEMENTS.btnNext) ELEMENTS.btnNext.addEventListener('click', nextPage);
        if (ELEMENTS.btnGo) ELEMENTS.btnGo.addEventListener('click', goToPage);
        var btnZoomIn  = document.getElementById('cg-btn-zoom-in');
        var btnZoomOut = document.getElementById('cg-btn-zoom-out');
        var btnZoomReset = document.getElementById('cg-btn-zoom-reset');
        if (btnZoomIn)  btnZoomIn.addEventListener('click', zoomIn);
        if (btnZoomOut) btnZoomOut.addEventListener('click', zoomOut);
        if (btnZoomReset) btnZoomReset.addEventListener('click', zoomReset);
        if (ELEMENTS.pageInput) {
            ELEMENTS.pageInput.addEventListener('keydown', function (e) {
                if (e.key === 'Enter') goToPage();
            });
        }
    }

    // ---- WASM Initialization ----

    function initWASM() {
        if (typeof WebAssembly === 'undefined') {
            showError('WebAssembly is not supported in this browser.');
            return;
        }

        // The Go WASM runtime sets up the 'go' global and calls the
        // run() method of the Go class. We use the standard wasm_exec.js
        // approach.
        if (typeof Go === 'undefined') {
            showError('Go WASM runtime (wasm_exec.js) not loaded.');
            return;
        }

        const go = new Go();

        WebAssembly.instantiateStreaming(fetch('cg-reader-wasm.wasm'), go.importObject)
            .then(function (result) {
                go.run(result.instance);
                // The Go program registers cgreader synchronously in main(),
                // but we add a small delay for the scheduler to settle.
                setTimeout(function () {
                    if (typeof cgreader === 'undefined' || typeof cgreader.new !== 'function') {
                        // Retry a few times.
                        var retries = 0;
                        var interval = setInterval(function () {
                            retries++;
                            if (typeof cgreader !== 'undefined' && typeof cgreader.new === 'function') {
                                clearInterval(interval);
                                goReady = true;
                                onReady();
                            } else if (retries > 20) {
                                clearInterval(interval);
                                showError('WASM module initialized but cgreader API not found.');
                            }
                        }, 100);
                    } else {
                        goReady = true;
                        onReady();
                    }
                }, 50);
            })
            .catch(function (err) {
                showError('Failed to initialize WASM: ' + err.message);
            });
    }

    function onReady() {
        console.log('[cg-reader] WASM ready, readerID=' + readerID);
        hideError();

        // If a comic URL is configured, load it.
        if (comicURL) {
            loadFromURL(comicURL, comicURL.split('/').pop());
        }
    }

    // ---- Bootstrap ----

    window.addEventListener('resize', function() { if (ELEMENTS.pageImg && ELEMENTS.pageImg.naturalWidth) { fitShellToImage(); } });

    function init() {
        readConfig();
        setupButtons();
        setupFileHandling();
        // Defer WASM init so it doesn't block the page load event.
        // The 12MB .wasm fetch+compile can trigger Chrome's [Violation] warning.
        setTimeout(initWASM, 0);
    }

    // Start when DOM is ready.
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
