//go:build js && wasm

package cgreaderwasm

import (
	"fmt"
	"syscall/js"
)

// JSExports registers all ComicReader methods on the global JavaScript object
// under the namespace "cgreader". Call this from your WASM main() function.
//
// Exposed JS API:
//
//	cgreader.new(passwordURL?: string, cacheEnabled?: bool) -> readerID
//	cgreader.openArchive(readerID, data: Uint8Array, filename: string) -> Promise<error?>
//	cgreader.pageCount(readerID) -> number
//	cgreader.currentPage(readerID) -> number
//	cgreader.getPage(readerID, page: number) -> Uint8Array
//	cgreader.getCurrentPage(readerID) -> Uint8Array
//	cgreader.next(readerID) -> bool
//	cgreader.prev(readerID) -> bool
//	cgreader.setCurrentPage(readerID, page: number) -> error?
//	cgreader.title(readerID) -> string
//	cgreader.filename(readerID) -> string
//	cgreader.setPasswordURL(readerID, url: string)
//	cgreader.close(readerID)
var (
	readers   = make(map[int]*ComicReader)
	nextID    = 0
	jsReaders = js.Value{}
)

// RegisterJS sets up the JavaScript API. Call once from main().
func RegisterJS() {
	cg := js.Global().Get("cgreader")
	if cg.IsUndefined() {
		cg = js.ValueOf(map[string]interface{}{})
		js.Global().Set("cgreader", cg)
	}
	// If already set up (e.g., object created in JS), use it.
	if cg.Type() == js.TypeObject {
		jsReaders = cg
	} else {
		jsReaders = js.ValueOf(map[string]interface{}{})
		js.Global().Set("cgreader", jsReaders)
	}

	jsReaders.Set("new", js.FuncOf(jsNew))
	jsReaders.Set("openArchive", js.FuncOf(jsOpenArchive))
	jsReaders.Set("pageCount", js.FuncOf(jsPageCount))
	jsReaders.Set("currentPage", js.FuncOf(jsCurrentPage))
	jsReaders.Set("getPage", js.FuncOf(jsGetPage))
	jsReaders.Set("getCurrentPage", js.FuncOf(jsGetCurrentPage))
	jsReaders.Set("next", js.FuncOf(jsNext))
	jsReaders.Set("prev", js.FuncOf(jsPrev))
	jsReaders.Set("setCurrentPage", js.FuncOf(jsSetCurrentPage))
	jsReaders.Set("title", js.FuncOf(jsTitle))
	jsReaders.Set("filename", js.FuncOf(jsFilename))
	jsReaders.Set("setPasswordURL", js.FuncOf(jsSetPasswordURL))
	jsReaders.Set("close", js.FuncOf(jsClose))
}

// ---- JS Function Implementations ----

func jsNew(this js.Value, args []js.Value) interface{} {
	passwordURL := ""
	cacheEnabled := true
	if len(args) > 0 && args[0].Type() == js.TypeString {
		passwordURL = args[0].String()
	}
	if len(args) > 1 && args[1].Type() == js.TypeBoolean {
		cacheEnabled = args[1].Bool()
	}

	reader := New(
		WithPasswordURL(passwordURL),
		WithCacheEnabled(cacheEnabled),
	)

	nextID++
	id := nextID
	readers[id] = reader
	return id
}

func jsOpenArchive(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return js.ValueOf("openArchive requires (readerID, data, filename)")
	}

	id := args[0].Int()
	reader, ok := readers[id]
	if !ok {
		return js.ValueOf(fmt.Sprintf("reader %d not found", id))
	}

	// Convert Uint8Array to []byte.
	data := make([]byte, args[1].Length())
	js.CopyBytesToGo(data, args[1])

	filename := args[2].String()

	// Return a Promise since OpenArchive may fetch password asynchronously.
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			err := reader.OpenArchive(data, filename)
			if err != nil {
				reject.Invoke(js.ValueOf(err.Error()))
			} else {
				resolve.Invoke(js.ValueOf(nil))
			}
		}()

		return nil
	})
	return js.Global().Get("Promise").New(handler)
}

func jsPageCount(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return 0
	}
	reader, ok := readers[args[0].Int()]
	if !ok {
		return 0
	}
	return reader.PageCount()
}

func jsCurrentPage(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return 0
	}
	reader, ok := readers[args[0].Int()]
	if !ok {
		return 0
	}
	return reader.CurrentPage()
}

func jsGetPage(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(fmt.Errorf("getPage requires (readerID, page)"))
	}
	reader, ok := readers[args[0].Int()]
	if !ok {
		return js.ValueOf(fmt.Errorf("reader %d not found", args[0].Int()))
	}

	data, err := reader.GetPage(args[1].Int())
	if err != nil {
		return js.ValueOf(fmt.Errorf("getPage: %w", err))
	}

	// Return as Uint8Array.
	dst := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(dst, data)
	return dst
}

func jsGetCurrentPage(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(fmt.Errorf("getCurrentPage requires readerID"))
	}
	return jsGetPage(this, []js.Value{args[0], js.ValueOf(readers[args[0].Int()].CurrentPage())})
}

func jsNext(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return false
	}
	reader, ok := readers[args[0].Int()]
	if !ok {
		return false
	}
	return reader.Next()
}

func jsPrev(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return false
	}
	reader, ok := readers[args[0].Int()]
	if !ok {
		return false
	}
	return reader.Prev()
}

func jsSetCurrentPage(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf("setCurrentPage requires (readerID, page)")
	}
	reader, ok := readers[args[0].Int()]
	if !ok {
		return js.ValueOf(fmt.Sprintf("reader %d not found", args[0].Int()))
	}
	err := reader.SetCurrentPage(args[1].Int())
	if err != nil {
		return js.ValueOf(err.Error())
	}
	return nil
}

func jsTitle(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return ""
	}
	reader, ok := readers[args[0].Int()]
	if !ok {
		return ""
	}
	return reader.Title()
}

func jsFilename(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return ""
	}
	reader, ok := readers[args[0].Int()]
	if !ok {
		return ""
	}
	return reader.Filename()
}

func jsSetPasswordURL(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf("setPasswordURL requires (readerID, url)")
	}
	reader, ok := readers[args[0].Int()]
	if !ok {
		return js.ValueOf(fmt.Sprintf("reader %d not found", args[0].Int()))
	}
	reader.SetPasswordURL(args[1].String())
	return nil
}

func jsClose(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return nil
	}
	id := args[0].Int()
	if reader, ok := readers[id]; ok {
		reader.mu.Lock()
		if reader.archive != nil {
			reader.archive.Close()
		}
		reader.mu.Unlock()
		delete(readers, id)
	}
	return nil
}
