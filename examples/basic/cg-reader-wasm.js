// cg-reader-wasm.js — JavaScript interop for CG Reader WASM
// Handles WASM initialization, DOM updates, user interactions, fullscreen, auto-hide chrome,
// double page mode, manga mode, info overlay, and shortcuts modal.

(function () {
    'use strict';

    const SHELL_ID = 'cg-reader-shell';
    const ELEMENTS = {
        shell:          document.getElementById(SHELL_ID),
        title:          document.getElementById('cg-reader-title'),
        pageInd:        document.getElementById('cg-reader-page-indicator'),
        pageImg:        document.getElementById('cg-reader-page'),
        pageImg2:       document.getElementById('cg-reader-page-2'),
        pageContainer:  document.getElementById('cg-page-container'),
        dropZone:       document.getElementById('drop-zone'),
        fileInput:      document.getElementById('file-input'),
        browseLink:     document.getElementById('browse-link'),
        btnPrev:        document.getElementById('cg-btn-prev'),
        btnNext:        document.getElementById('cg-btn-next'),
        btnGo:          document.getElementById('cg-btn-go'),
        pageInput:      document.getElementById('cg-page-input'),
        pageTotal:      document.getElementById('cg-page-total'),
        errorBox:       document.getElementById('cg-reader-error'),
        infoOverlay:    document.getElementById('cg-info-overlay'),
        infoPages:      document.getElementById('cg-info-pages'),
        infoTime:       document.getElementById('cg-info-time'),
        shortcutsModal: document.getElementById('cg-shortcuts-modal'),
        fsEnterIcon:    document.querySelector('.cg-fs-enter'),
        fsExitIcon:     document.querySelector('.cg-fs-exit'),
    };

    let readerID        = null;
    let zoomLevel       = 100;
    let comicURL        = '';
    let passwordURL     = '';
    let goReady         = false;
    let doublePageMode  = false;
    let mangaMode       = false;
    let infoVisible     = false;
    let infoClockID     = null;

    // ---- Toolbar button active state helpers ----

    function setToolActive(btn, active) {
        if (!btn) return;
        if (active) {
            btn.classList.add('text-cg-accent');
            btn.classList.remove('text-cg-muted');
        } else {
            btn.classList.add('text-cg-muted');
            btn.classList.remove('text-cg-accent');
        }
    }

    function refreshToolStates() {
        setToolActive(ELEMENTS.btnDouble, doublePageMode);
        ELEMENTS.btnManga.disabled = !doublePageMode;
        setToolActive(ELEMENTS.btnManga, mangaMode && doublePageMode);
        setToolActive(ELEMENTS.btnInfo, infoVisible);
    }

    // ---- Fullscreen auto-hide ----

    let fsIdleTimer = null;
    const FS_IDLE_DELAY = 3000;

    function resetFsIdleTimer() {
        if (!document.fullscreenElement || ELEMENTS.shell !== document.fullscreenElement) return;
        ELEMENTS.shell.classList.remove('cg-fullscreen-idle');
        if (fsIdleTimer) clearTimeout(fsIdleTimer);
        fsIdleTimer = setTimeout(function () {
            if (document.fullscreenElement === ELEMENTS.shell) {
                ELEMENTS.shell.classList.add('cg-fullscreen-idle');
            }
        }, FS_IDLE_DELAY);
    }

    function clearFsIdleTimer() {
        if (fsIdleTimer) {
            clearTimeout(fsIdleTimer);
            fsIdleTimer = null;
        }
        if (ELEMENTS.shell) {
            ELEMENTS.shell.classList.remove('cg-fullscreen-idle');
        }
    }

    // ---- Read configuration ----

    function readConfig() {
        var shell = ELEMENTS.shell;
        if (!shell) return;
        comicURL = shell.getAttribute('data-comic-url') || '';
        passwordURL = shell.getAttribute('data-password-url') || '';
    }

    // ---- Error / Spinner ----

    function showSpinner(text) {
        var s = document.getElementById('cg-spinner');
        var t = document.getElementById('cg-spinner-text');
        if (s) { s.style.display = 'flex'; }
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

    // ---- Info overlay ----

    function updateInfoOverlay() {
        if (!infoVisible) return;
        ELEMENTS.infoTime.textContent = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        if (readerID !== null) {
            try {
                var current = cgreader.currentPage(readerID);
                var total = cgreader.pageCount(readerID);
                if (total > 1 && doublePageMode && current > 0) {
                    var right = Math.min(current + 1, total - 1);
                    ELEMENTS.infoPages.textContent = 'Page ' + (current + 1) + '\u2013' + (right + 1) + ' / ' + total;
                } else {
                    ELEMENTS.infoPages.textContent = 'Page ' + (current + 1) + ' / ' + total;
                }
            } catch(e) { /* ignore */ }
        }
    }

    function startInfoClock() {
        updateInfoOverlay();
        if (infoClockID) clearInterval(infoClockID);
        infoClockID = setInterval(updateInfoOverlay, 30000);
    }

    function stopInfoClock() {
        if (infoClockID) {
            clearInterval(infoClockID);
            infoClockID = null;
        }
    }

    function toggleInfo() {
        infoVisible = !infoVisible;
        if (infoVisible) {
            ELEMENTS.infoOverlay.style.display = 'flex';
            startInfoClock();
        } else {
            ELEMENTS.infoOverlay.style.display = 'none';
            stopInfoClock();
        }
        refreshToolStates();
    }

    // ---- Shortcuts modal ----

    function openShortcutsModal() {
        if (ELEMENTS.shortcutsModal) ELEMENTS.shortcutsModal.style.display = 'flex';
    }
    function closeShortcutsModal() {
        if (ELEMENTS.shortcutsModal) ELEMENTS.shortcutsModal.style.display = 'none';
    }
    function isShortcutsModalOpen() {
        return ELEMENTS.shortcutsModal && ELEMENTS.shortcutsModal.style.display !== 'none';
    }

    // ---- Double page helpers ----

    function renderToImage(img, data) {
        if (!data || !(data instanceof Uint8Array) || data.length === 0) {
            img.style.display = 'none';
            return;
        }
        var blob = new Blob([data], { type: 'image/jpeg' });
        var url = URL.createObjectURL(blob);
        var old = img.src;
        if (old && old.startsWith('blob:')) URL.revokeObjectURL(old);
        img.src = url;
        img.style.display = 'block';
    }

    function renderSinglePage(current) {
        if (ELEMENTS.pageImg2) ELEMENTS.pageImg2.style.display = 'none';
        if (ELEMENTS.pageContainer) {
            ELEMENTS.pageContainer.style.flexDirection = '';
            ELEMENTS.pageContainer.style.width = '';
            ELEMENTS.pageContainer.style.height = '';
        }
        if (ELEMENTS.pageImg) {
            ELEMENTS.pageImg.style.maxWidth = '';
            ELEMENTS.pageImg.style.maxHeight = '';
            ELEMENTS.pageImg.style.width = '';
            ELEMENTS.pageImg.style.height = '';
        }
        if (ELEMENTS.pageImg2) {
            ELEMENTS.pageImg2.style.maxWidth = '';
            ELEMENTS.pageImg2.style.maxHeight = '';
            ELEMENTS.pageImg2.style.width = '';
            ELEMENTS.pageImg2.style.height = '';
        }

        try {
            var data = cgreader.getPage(readerID, current);
            renderToImage(ELEMENTS.pageImg, data);
        } catch(e) {
            showError('Render error: ' + e.message);
            return;
        }
        ELEMENTS.pageImg.onload = function() {
            updateZoom();
            fitShellToImage();
        };
    }

    function renderDoublePage(current, total) {
        if (ELEMENTS.pageContainer) {
            ELEMENTS.pageContainer.style.flexDirection = 'row';
            ELEMENTS.pageContainer.style.width = '100%';
            ELEMENTS.pageContainer.style.height = '100%';
        }

        var leftIdx, rightIdx;
        if (mangaMode) {
            leftIdx  = Math.min(current + 1, total - 1);
            rightIdx = current;
        } else {
            leftIdx  = current;
            rightIdx = Math.min(current + 1, total - 1);
        }

        if (ELEMENTS.pageImg) {
            ELEMENTS.pageImg.style.width = '50%';
            ELEMENTS.pageImg.style.height = '100%';
            ELEMENTS.pageImg.style.maxWidth = '50%';
            ELEMENTS.pageImg.style.maxHeight = '100%';
        }
        if (ELEMENTS.pageImg2) {
            ELEMENTS.pageImg2.style.width = '50%';
            ELEMENTS.pageImg2.style.height = '100%';
            ELEMENTS.pageImg2.style.maxWidth = '50%';
            ELEMENTS.pageImg2.style.maxHeight = '100%';
        }

        try {
            var leftData = cgreader.getPage(readerID, leftIdx);
            renderToImage(ELEMENTS.pageImg, leftData);
        } catch(e) {
            showError('Render error: ' + e.message);
            return;
        }

        try {
            var rightData = cgreader.getPage(readerID, rightIdx);
            renderToImage(ELEMENTS.pageImg2, rightData);
        } catch(e) {
            showError('Render error: ' + e.message);
            return;
        }

        ELEMENTS.pageImg.onload = function() {
            updateZoom();
            fitShellToImage();
        };
    }

    // ---- UI Updates ----

    function updateUI() {
        if (!goReady || readerID === null) return;

        try {
            var pageCount = cgreader.pageCount(readerID);
            var current   = cgreader.currentPage(readerID);
            var title     = cgreader.title(readerID);

            if (ELEMENTS.title) {
                ELEMENTS.title.textContent = title || 'CG Reader';
            }
            if (pageCount > 1 && doublePageMode && current > 0) {
                var right = Math.min(current + 1, pageCount - 1);
                ELEMENTS.pageInd.textContent = (current + 1) + '\u2013' + (right + 1) + ' / ' + pageCount;
            } else if (ELEMENTS.pageInd) {
                ELEMENTS.pageInd.textContent = (current + 1) + ' / ' + pageCount;
            }
            if (ELEMENTS.pageInput) {
                ELEMENTS.pageInput.value = current + 1;
                ELEMENTS.pageInput.max = pageCount;
            }
            if (ELEMENTS.pageTotal) {
                ELEMENTS.pageTotal.textContent = pageCount;
            }

            var hasPages = pageCount > 0;
            [ELEMENTS.btnPrev, ELEMENTS.btnNext, ELEMENTS.btnGo, ELEMENTS.pageInput].forEach(function (el) {
                if (el) el.disabled = !hasPages;
            });
            if (ELEMENTS.btnPrev) ELEMENTS.btnPrev.disabled = !hasPages || current <= 0;
            if (ELEMENTS.btnNext) {
                if (doublePageMode && current > 0) {
                    ELEMENTS.btnNext.disabled = !hasPages || current + 2 >= pageCount;
                } else {
                    ELEMENTS.btnNext.disabled = !hasPages || current >= pageCount - 1;
                }
            }

            updateInfoOverlay();
        } catch (e) {
            showError('UI update error: ' + e.message);
        }
    }

    function fitShellToImage() {
        var img = ELEMENTS.pageImg;
        if (!img || !img.naturalWidth || !img.naturalHeight) return;
        var shell = ELEMENTS.shell;
        var maxW = window.innerWidth * 0.95;
        var maxH = window.innerHeight * 0.85;
        var ratio = img.naturalWidth / img.naturalHeight;
        var w, h;
        if (ratio > maxW / maxH) {
            w = maxW;
            h = w / ratio;
        } else {
            h = maxH;
            w = h * ratio;
        }
        shell.style.maxWidth = Math.floor(w + 48) + 'px';
        var vp = document.getElementById('cg-reader-viewport');
        if (vp) {
            vp.style.maxHeight = Math.floor(h + 32) + 'px';
        }
    }

    function renderPage() {
        hideSpinner();
        if (!goReady || readerID === null) return;

        var pageCount = cgreader.pageCount(readerID);
        if (pageCount === 0) return;

        var current = cgreader.currentPage(readerID);

        if (doublePageMode && current > 0 && pageCount > 1) {
            renderDoublePage(current, pageCount);
        } else {
            renderSinglePage(current);
        }
        if (ELEMENTS.dropZone) ELEMENTS.dropZone.style.display = 'none';
    }

    // ---- Navigation ----

    function goToPage() {
        if (!goReady || readerID === null) return;
        hideError();

        var page = parseInt(ELEMENTS.pageInput.value, 10) - 1;
        if (isNaN(page)) return;

        try {
            var err = cgreader.setCurrentPage(readerID, page);
            if (err) { showError(err); return; }
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
            var pageCount = cgreader.pageCount(readerID);
            var current = cgreader.currentPage(readerID);

            if (doublePageMode) {
                if (current === 0 && pageCount > 1) {
                    cgreader.setCurrentPage(readerID, 1);
                } else if (current + 2 < pageCount) {
                    cgreader.setCurrentPage(readerID, current + 2);
                } else {
                    return;
                }
            } else {
                if (!cgreader.next(readerID)) return;
            }
            updateUI();
            renderPage();
        } catch (e) {
            showError(e.message);
        }
    }

    function prevPage() {
        if (!goReady || readerID === null) return;
        hideError();

        try {
            var current = cgreader.currentPage(readerID);

            if (doublePageMode) {
                if (current <= 0) return;
                if (current === 1) {
                    cgreader.setCurrentPage(readerID, 0);
                } else {
                    cgreader.setCurrentPage(readerID, Math.max(0, current - 2));
                }
            } else {
                if (!cgreader.prev(readerID)) return;
            }
            updateUI();
            renderPage();
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

        if (readerID !== null) {
            try { cgreader.close(readerID); } catch (e) { /* ignore */ }
            readerID = null;
        }
        readerID = cgreader.new(passwordURL, true);

        var uint8 = new Uint8Array(data);
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
        reader.onload = function (e) { openArchive(e.target.result, file.name); };
        reader.onerror = function () { showError('Failed to read file.'); };
        reader.readAsArrayBuffer(file);
    }

    // ---- Double page / Manga toggles ----

    function toggleDoublePage() {
        doublePageMode = !doublePageMode;
        if (!doublePageMode) mangaMode = false;
        refreshToolStates();
        renderPage();
        updateUI();
    }

    function toggleMangaMode() {
        if (!doublePageMode) return;
        mangaMode = !mangaMode;
        refreshToolStates();
        renderPage();
        updateUI();
    }

    // ---- File Drop / Browse / Keyboard ----

    function setupFileHandling() {
        if (ELEMENTS.browseLink && ELEMENTS.fileInput) {
            ELEMENTS.browseLink.addEventListener('click', function (e) {
                e.stopPropagation();
                ELEMENTS.fileInput.click();
            });
            ELEMENTS.fileInput.addEventListener('change', function () {
                if (this.files && this.files[0]) handleFile(this.files[0]);
            });
        }
        if (ELEMENTS.dropZone) {
            ELEMENTS.dropZone.addEventListener('click', function () {
                if (ELEMENTS.fileInput) ELEMENTS.fileInput.click();
            });
        }

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

        document.addEventListener('keydown', function (e) {
            var tag = e.target.tagName;
            if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return;

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
                    if (readerID !== null) {
                        var total = cgreader.pageCount(readerID);
                        if (ELEMENTS.pageInput && total > 0) {
                            ELEMENTS.pageInput.value = String(total);
                            goToPage();
                        }
                    }
                    break;
                case 'f':
                case 'F':
                    e.preventDefault();
                    toggleFullscreen();
                    break;
                case 'd':
                case 'D':
                    e.preventDefault();
                    toggleDoublePage();
                    break;
                case 'm':
                case 'M':
                    e.preventDefault();
                    toggleMangaMode();
                    break;
                case 'i':
                case 'I':
                    e.preventDefault();
                    toggleInfo();
                    break;
                case 'Escape':
                    e.preventDefault();
                    if (isShortcutsModalOpen()) {
                        closeShortcutsModal();
                    } else {
                        exitFullscreen();
                    }
                    break;
                case 'r':
                case 'R':
                    e.preventDefault();
                    resumeLastPage();
                    break;
                case '?':
                    e.preventDefault();
                    if (isShortcutsModalOpen()) {
                        closeShortcutsModal();
                    } else {
                        openShortcutsModal();
                    }
                    break;
            }
        });
    }

    // ---- Button Events ----

    function updateZoom() {
        var img = ELEMENTS.pageImg;
        var vp = document.getElementById('cg-reader-viewport');
        if (!img) return;
        if (zoomLevel === 100) {
            img.style.transform = '';
            img.style.transformOrigin = '';
            img.style.maxWidth = '100%';
            img.style.maxHeight = '100%';
            if (vp) vp.style.overflow = 'hidden';
        } else {
            img.style.transform = 'scale(' + (zoomLevel / 100) + ')';
            img.style.transformOrigin = 'top left';
            img.style.maxWidth = 'none';
            img.style.maxHeight = 'none';
            if (vp) vp.style.overflow = 'auto';
        }
        var label = document.getElementById('cg-zoom-level');
        if (label) label.textContent = zoomLevel + '%';
    }
    function zoomIn()  { zoomLevel = Math.min(300, zoomLevel + 25); updateZoom(); }
    function zoomOut() { zoomLevel = Math.max(25, zoomLevel - 25); updateZoom(); }
    function zoomReset() { zoomLevel = 100; updateZoom(); }

    // ---- Fullscreen ----

    function toggleFullscreen() {
        var shell = ELEMENTS.shell;
        if (!shell) return;
        if (document.fullscreenElement) {
            document.exitFullscreen();
        } else {
            saveLastPage();
            if (shell.requestFullscreen) shell.requestFullscreen();
        }
    }

    function exitFullscreen() {
        if (document.fullscreenElement) document.exitFullscreen();
    }

    // ---- Resume last page ----

    function saveLastPage() {
        if (readerID === null) return;
        try {
            localStorage.setItem('cg-reader-last-page', cgreader.currentPage(readerID));
        } catch (e) { /* ignore */ }
    }

    function resumeLastPage() {
        try {
            var saved = localStorage.getItem('cg-reader-last-page');
            if (saved !== null && readerID !== null) {
                var page = parseInt(saved, 10);
                if (!cgreader.setCurrentPage(readerID, page)) {
                    updateUI();
                    renderPage();
                }
            }
        } catch (e) { /* ignore */ }
    }

    var _origPrevPage = prevPage;
    var _origNextPage = nextPage;
    var _origGoToPage = goToPage;
    prevPage = function() { saveLastPage(); _origPrevPage(); };
    nextPage = function() { saveLastPage(); _origNextPage(); };
    goToPage = function() { saveLastPage(); _origGoToPage(); };

    function setupButtons() {
        if (ELEMENTS.btnPrev)  ELEMENTS.btnPrev.addEventListener('click', prevPage);
        if (ELEMENTS.btnNext)  ELEMENTS.btnNext.addEventListener('click', nextPage);
        if (ELEMENTS.btnGo)    ELEMENTS.btnGo.addEventListener('click', goToPage);

        var btnZoomIn   = document.getElementById('cg-btn-zoom-in');
        var btnZoomOut  = document.getElementById('cg-btn-zoom-out');
        var btnZoomReset = document.getElementById('cg-btn-zoom-reset');
        if (btnZoomIn)   btnZoomIn.addEventListener('click', zoomIn);
        if (btnZoomOut)  btnZoomOut.addEventListener('click', zoomOut);
        if (btnZoomReset) btnZoomReset.addEventListener('click', zoomReset);

        if (ELEMENTS.pageInput) {
            ELEMENTS.pageInput.addEventListener('keydown', function (e) {
                if (e.key === 'Enter') goToPage();
            });
        }

        // Toolbar button refs — cached after DOM ready
        ELEMENTS.btnInfo      = document.getElementById('cg-btn-info');
        ELEMENTS.btnDouble    = document.getElementById('cg-btn-double');
        ELEMENTS.btnManga     = document.getElementById('cg-btn-manga');
        ELEMENTS.btnFullscreen = document.getElementById('cg-btn-fullscreen');
        ELEMENTS.btnShortcuts = document.getElementById('cg-btn-shortcuts');
        ELEMENTS.btnCloseModal = document.getElementById('cg-btn-close-modal');

        if (ELEMENTS.btnInfo)      ELEMENTS.btnInfo.addEventListener('click', toggleInfo);
        if (ELEMENTS.btnDouble)    ELEMENTS.btnDouble.addEventListener('click', toggleDoublePage);
        if (ELEMENTS.btnManga)     ELEMENTS.btnManga.addEventListener('click', toggleMangaMode);
        if (ELEMENTS.btnFullscreen) ELEMENTS.btnFullscreen.addEventListener('click', toggleFullscreen);
        if (ELEMENTS.btnShortcuts) ELEMENTS.btnShortcuts.addEventListener('click', function () {
            if (isShortcutsModalOpen()) closeShortcutsModal(); else openShortcutsModal();
        });
        if (ELEMENTS.btnCloseModal) ELEMENTS.btnCloseModal.addEventListener('click', closeShortcutsModal);

        if (ELEMENTS.shortcutsModal) {
            ELEMENTS.shortcutsModal.addEventListener('click', function (e) {
                if (e.target === ELEMENTS.shortcutsModal) closeShortcutsModal();
            });
        }

        refreshToolStates();
    }

    // ---- WASM Initialization ----

    function initWASM() {
        if (typeof WebAssembly === 'undefined') {
            showError('WebAssembly is not supported in this browser.');
            return;
        }
        if (typeof Go === 'undefined') {
            showError('Go WASM runtime (wasm_exec.js) not loaded.');
            return;
        }

        var go = new Go();
        WebAssembly.instantiateStreaming(fetch('cg-reader-wasm.wasm'), go.importObject)
            .then(function (result) {
                go.run(result.instance);
                setTimeout(function () {
                    if (typeof cgreader === 'undefined' || typeof cgreader.new !== 'function') {
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
        if (comicURL) loadFromURL(comicURL, comicURL.split('/').pop());
    }

    // ---- Bootstrap ----

    document.addEventListener('fullscreenchange', function() {
        var shell = ELEMENTS.shell;
        var isFS = document.fullscreenElement === shell;
        if (ELEMENTS.fsEnterIcon) ELEMENTS.fsEnterIcon.classList.toggle('hidden', isFS);
        if (ELEMENTS.fsExitIcon)  ELEMENTS.fsExitIcon.classList.toggle('hidden', !isFS);
        if (isFS) {
            shell.classList.add('cg-fullscreen-active');
            fitShellToImage();
            resetFsIdleTimer();
        } else {
            clearFsIdleTimer();
            shell.classList.remove('cg-fullscreen-active', 'cg-fullscreen-idle');
            fitShellToImage();
            saveLastPage();
        }
    });

    if (ELEMENTS.shell) {
        ELEMENTS.shell.addEventListener('mousemove', function () {
            if (document.fullscreenElement === ELEMENTS.shell) resetFsIdleTimer();
        });
        ELEMENTS.shell.addEventListener('touchstart', function () {
            if (document.fullscreenElement === ELEMENTS.shell) resetFsIdleTimer();
        });
    }

    window.addEventListener('resize', function() {
        if (ELEMENTS.pageImg && ELEMENTS.pageImg.naturalWidth) fitShellToImage();
    });

    function init() {
        readConfig();
        setupButtons();
        setupFileHandling();
        setTimeout(initWASM, 0);
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
