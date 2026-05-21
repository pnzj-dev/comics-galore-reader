### 2. Example Templ Component (`reader.templ`)

```templ
package components

templ ReaderShell() {
	<div class="comic-reader h-screen flex flex-col bg-zinc-950 text-white">
		<!-- Header -->
		<header class="bg-zinc-900 border-b border-zinc-800 p-4 flex items-center justify-between">
			<div class="flex items-center gap-4">
				<h1 class="text-xl font-bold" id="comic-title">Loading comic...</h1>
			</div>
			<div class="flex items-center gap-6">
				<div class="text-lg font-mono" id="page-counter">0 / 0</div>
				
				<div class="flex gap-2">
					<button 
					class="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded"
					onclick="prevPage()">← Prev</button>
					<button 
					class="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded"
					onclick="nextPage()">Next →</button>
				</div>
			</div>
		</header>

		<!-- Main Viewer -->
		<div class="flex-1 flex items-center justify-center bg-black overflow-hidden" id="viewer">
			<img 
				id="comic-page" 
				class="max-h-full max-w-full object-contain cursor-pointer"
				onclick="nextPage()"
				alt="Comic page"
			/>
		</div>

		<!-- Controls -->
		<footer class="bg-zinc-900 border-t border-zinc-800 p-3">
			<input 
				type="range" 
				id="page-slider"
				min="0" 
				value="0"
				oninput="goToPage(this.value)"
				class="w-full"
			/>
		</footer>
	</div>
}
```

-----

### 3. How to Call WASM from Templ + HTMX / JavaScript

Create `static/reader.js`:

```javascript
let reader; // Will hold the WASM ComicReader instance

async function initWasm() {
    // Load your WASM module
    const go = new Go();
    const wasm = await WebAssembly.instantiateStreaming(
        fetch("/comic.wasm"), 
        go.importObject
    );
    go.run(wasm.instance);
    
    // Initialize reader with default password URL
    reader = comicwasm.NewComicReader(
        comicwasm.WithPasswordURL("/api/comic-password")
    );
}

// Override password URL at runtime
function setPasswordUrl(newUrl) {
    if (reader) reader.SetPasswordURL(newUrl);
}

// Open a file (called from file input)
async function openComic(file) {
    const arrayBuffer = await file.arrayBuffer();
    const bytes = new Uint8Array(arrayBuffer);
    
    try {
        await reader.OpenArchive(bytes, file.name);
        
        document.getElementById("comic-title").textContent = reader.Title();
        updatePageInfo();
        loadCurrentPage();
    } catch (e) {
        console.error(e);
        // Handle password error, etc.
    }
}

async function loadCurrentPage() {
    const imgData = await reader.GetPage(reader.CurrentPage());
    const blob = new Blob([imgData], { type: "image/jpeg" }); // or png/webp
    const url = URL.createObjectURL(blob);
    
    const img = document.getElementById("comic-page");
    img.src = url;
}

function updatePageInfo() {
    const current = reader.CurrentPage() + 1;
    const total = reader.PageCount();
    document.getElementById("page-counter").textContent = `${current} / ${total}`;
    document.getElementById("page-slider").max = total - 1;
    document.getElementById("page-slider").value = reader.CurrentPage();
}

// Navigation
function nextPage() {
    if (reader.Next()) {
        updatePageInfo();
        loadCurrentPage();
    }
}

function prevPage() {
    if (reader.Prev()) {
        updatePageInfo();
        loadCurrentPage();
    }
}

function goToPage(page) {
    reader.SetCurrentPage(parseInt(page));
    updatePageInfo();
    loadCurrentPage();
}
```

**Usage in your main Templ page**:

```templ
@components.ReaderShell()
<script src="/reader.js"></script>
<script>
    window.onload = initWasm;
</script>
```
