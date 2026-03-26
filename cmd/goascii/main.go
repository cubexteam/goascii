// goascii — конвертер изображений в ASCII-арт с веб-интерфейсом
// Автор: SantianDev | https://github.com/cubexteam/goascii
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cubexteam/goascii/internal/charset"
	"github.com/cubexteam/goascii/internal/converter"
)

// convertRequest — тело запроса от фронтенда
type convertRequest struct {
	Charset string `json:"charset"`
	Width   int    `json:"width"`
	Invert  bool   `json:"invert"`
	Colored bool   `json:"colored"`
}

// convertResponse — ответ сервера с результатом
type convertResponse struct {
	HTML  string `json:"html"`
	Plain string `json:"plain"`
	Error string `json:"error,omitempty"`
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/convert", handleConvert)

	addr := "127.0.0.1:7821"
	url := "http://" + addr

	fmt.Println("goascii by SantianDev")
	fmt.Println("https://github.com/cubexteam/goascii")
	fmt.Println()
	fmt.Printf("Server running at %s\n", url)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Открываем браузер автоматически
	openBrowser(url)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// openBrowser открывает браузер с нужным URL на любой ОС
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

// handleConvert принимает multipart/form-data с изображением и параметрами
func handleConvert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	// Лимит загрузки 20MB
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeError(w, "Failed to parse form: "+err.Error())
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		writeError(w, "No image provided")
		return
	}
	defer file.Close()

	imgData, err := io.ReadAll(file)
	if err != nil {
		writeError(w, "Failed to read image")
		return
	}

	// Читаем параметры из формы
	cs := r.FormValue("charset")
	if cs == "" {
		cs = "detailed"
	}

	width := 160
	if v := r.FormValue("width"); v != "" {
		fmt.Sscanf(v, "%d", &width)
	}
	if width < 20 {
		width = 20
	}
	if width > 300 {
		width = 300
	}

	invert := r.FormValue("invert") == "true"
	colored := r.FormValue("colored") == "true"

	opts := converter.Options{
		Width:   width,
		Invert:  invert,
		Colored: colored,
		Charset: charset.Get(cs),
	}

	result, err := converter.ConvertBytes(imgData, opts)
	if err != nil {
		writeError(w, "Conversion failed: "+err.Error())
		return
	}

	// Собираем HTML и plain версии
	htmlOut := buildHTML(result)
	plainOut := buildPlain(result)

	resp := convertResponse{HTML: htmlOut, Plain: plainOut}
	json.NewEncoder(w).Encode(resp)
}

// buildHTML строит HTML с цветными span для каждого символа
func buildHTML(result *converter.Result) string {
	var sb strings.Builder
	for _, row := range result.Rows {
		for _, px := range row {
			ch := htmlEscape(px.Char)
			fmt.Fprintf(&sb, `<span style="color:rgb(%d,%d,%d)">%s</span>`, px.R, px.G, px.B, ch)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// buildPlain строит обычный текст без цвета
func buildPlain(result *converter.Result) string {
	var sb strings.Builder
	for _, row := range result.Rows {
		for _, px := range row {
			sb.WriteRune(px.Char)
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// htmlEscape экранирует спецсимволы HTML
func htmlEscape(r rune) string {
	switch r {
	case '<':
		return "&lt;"
	case '>':
		return "&gt;"
	case '&':
		return "&amp;"
	case ' ':
		return "&nbsp;"
	default:
		return string(r)
	}
}

func writeError(w http.ResponseWriter, msg string) {
	json.NewEncoder(w).Encode(convertResponse{Error: msg})
}

// handleIndex отдаёт HTML страницу интерфейса
func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

// HTML
const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>goascii</title>
<link href="https://fonts.googleapis.com/css2?family=Share+Tech+Mono&family=Syne:wght@400;700;800&display=swap" rel="stylesheet">
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

  :root {
    --bg: #0a0a0a;
    --surface: #111111;
    --border: #222222;
    --accent: #00ff88;
    --accent2: #00ccff;
    --text: #e8e8e8;
    --muted: #555;
    --danger: #ff4455;
    --font-mono: 'Share Tech Mono', monospace;
    --font-ui: 'Syne', sans-serif;
  }

  html, body {
    height: 100%;
    background: var(--bg);
    color: var(--text);
    font-family: var(--font-ui);
    overflow-x: hidden;
  }

  body::before {
    content: '';
    position: fixed;
    inset: 0;
    background-image: radial-gradient(circle, #1a1a1a 1px, transparent 1px);
    background-size: 28px 28px;
    pointer-events: none;
    z-index: 0;
  }

  .app {
    position: relative;
    z-index: 1;
    min-height: 100vh;
    display: grid;
    grid-template-rows: auto 1fr;
  }

  header {
    padding: 28px 40px;
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: rgba(10,10,10,0.9);
    backdrop-filter: blur(10px);
    position: sticky;
    top: 0;
    z-index: 100;
  }

  .logo {
    display: flex;
    align-items: baseline;
    gap: 10px;
  }

  .logo-text {
    font-size: 22px;
    font-weight: 800;
    letter-spacing: -0.5px;
    color: var(--accent);
  }

  .logo-sub {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--muted);
    letter-spacing: 0.05em;
  }

  .header-link {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--muted);
    text-decoration: none;
    border: 1px solid var(--border);
    padding: 6px 14px;
    border-radius: 4px;
    transition: all 0.2s;
  }
  .header-link:hover { color: var(--accent); border-color: var(--accent); }

  main {
    display: grid;
    grid-template-columns: 320px 1fr;
    gap: 0;
    height: calc(100vh - 73px);
  }

  .sidebar {
    border-right: 1px solid var(--border);
    padding: 28px 24px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 24px;
    background: rgba(10,10,10,0.5);
  }

  .section-label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.15em;
    text-transform: uppercase;
    color: var(--muted);
    margin-bottom: 10px;
  }

  .dropzone {
    border: 1.5px dashed var(--border);
    border-radius: 8px;
    padding: 32px 20px;
    text-align: center;
    cursor: pointer;
    transition: all 0.25s;
    background: rgba(255,255,255,0.015);
    position: relative;
  }

  .dropzone:hover, .dropzone.drag-over {
    border-color: var(--accent);
    background: rgba(0,255,136,0.04);
  }

  .dropzone-icon {
    font-size: 28px;
    margin-bottom: 10px;
    display: block;
    opacity: 0.5;
  }

  .dropzone-text {
    font-size: 13px;
    color: var(--muted);
    line-height: 1.6;
  }

  .dropzone-text strong {
    color: var(--accent);
    font-weight: 700;
  }

  .dropzone input[type=file] {
    position: absolute;
    inset: 0;
    opacity: 0;
    cursor: pointer;
  }

  .preview-img {
    width: 100%;
    max-height: 120px;
    object-fit: cover;
    border-radius: 6px;
    display: none;
    margin-top: 12px;
  }

  .control-group { display: flex; flex-direction: column; gap: 8px; }

  .charset-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6px;
  }

  .charset-btn {
    background: var(--surface);
    border: 1px solid var(--border);
    color: var(--text);
    font-family: var(--font-mono);
    font-size: 12px;
    padding: 9px 12px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
    text-align: left;
  }

  .charset-btn:hover { border-color: #333; background: #161616; }
  .charset-btn.active { border-color: var(--accent); color: var(--accent); background: rgba(0,255,136,0.06); }

  .charset-btn .cs-name { display: block; font-weight: bold; margin-bottom: 2px; }
  .charset-btn .cs-preview { display: block; font-size: 10px; color: var(--muted); overflow: hidden; white-space: nowrap; }

  .slider-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 6px;
  }

  .slider-val {
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--accent);
    min-width: 30px;
    text-align: right;
  }

  input[type=range] {
    width: 100%;
    height: 3px;
    -webkit-appearance: none;
    background: var(--border);
    border-radius: 2px;
    outline: none;
  }

  input[type=range]::-webkit-slider-thumb {
    -webkit-appearance: none;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    background: var(--accent);
    cursor: pointer;
    transition: transform 0.15s;
  }

  input[type=range]::-webkit-slider-thumb:hover { transform: scale(1.3); }

  .toggle-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 6px;
  }

  .toggle-label { font-size: 13px; }
  .toggle-desc { font-family: var(--font-mono); font-size: 10px; color: var(--muted); display: block; margin-top: 1px; }

  .toggle {
    width: 38px;
    height: 20px;
    background: var(--border);
    border-radius: 10px;
    cursor: pointer;
    position: relative;
    transition: background 0.2s;
    border: none;
    flex-shrink: 0;
  }

  .toggle::after {
    content: '';
    position: absolute;
    width: 14px;
    height: 14px;
    background: #444;
    border-radius: 50%;
    top: 3px;
    left: 3px;
    transition: all 0.2s;
  }

  .toggle.on { background: rgba(0,255,136,0.2); }
  .toggle.on::after { left: 21px; background: var(--accent); }

  .convert-btn {
    width: 100%;
    padding: 14px;
    background: var(--accent);
    color: #000;
    border: none;
    border-radius: 8px;
    font-family: var(--font-ui);
    font-size: 14px;
    font-weight: 800;
    letter-spacing: 0.05em;
    cursor: pointer;
    transition: all 0.2s;
    text-transform: uppercase;
    position: relative;
    overflow: hidden;
  }

  .convert-btn:hover { background: #00e87a; transform: translateY(-1px); box-shadow: 0 4px 20px rgba(0,255,136,0.25); }
  .convert-btn:active { transform: translateY(0); }
  .convert-btn:disabled { background: var(--border); color: var(--muted); cursor: not-allowed; transform: none; box-shadow: none; }

  .convert-btn .btn-spinner {
    display: none;
    width: 16px;
    height: 16px;
    border: 2px solid rgba(0,0,0,0.3);
    border-top-color: #000;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
    margin: 0 auto;
  }

  .convert-btn.loading .btn-text { display: none; }
  .convert-btn.loading .btn-spinner { display: block; }

  @keyframes spin { to { transform: rotate(360deg); } }

  .output-panel {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--bg);
  }

  .output-toolbar {
    padding: 14px 24px;
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    gap: 12px;
    background: rgba(10,10,10,0.8);
    flex-shrink: 0;
  }

  .output-tabs {
    display: flex;
    gap: 4px;
    background: var(--surface);
    padding: 3px;
    border-radius: 6px;
    border: 1px solid var(--border);
  }

  .tab-btn {
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 5px 14px;
    border: none;
    background: none;
    color: var(--muted);
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.15s;
  }

  .tab-btn.active { background: var(--border); color: var(--text); }

  .toolbar-spacer { flex: 1; }

  .icon-btn {
    background: var(--surface);
    border: 1px solid var(--border);
    color: var(--muted);
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 6px 14px;
    border-radius: 5px;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .icon-btn:hover { color: var(--text); border-color: #333; }
  .icon-btn.accent:hover { color: var(--accent); border-color: var(--accent); }

  .output-scroll {
    flex: 1;
    overflow: auto;
    padding: 20px 24px;
  }

  .ascii-out {
    font-family: var(--font-mono);
    font-size: 7px;
    line-height: 1.15;
    white-space: pre;
    letter-spacing: 0.02em;
    display: none;
  }

  .ascii-out.visible { display: block; }

  .empty-state {
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    opacity: 0.3;
  }

  .empty-state-art {
    font-family: var(--font-mono);
    font-size: 10px;
    line-height: 1.3;
    color: var(--muted);
    text-align: center;
    white-space: pre;
  }

  .empty-state-text {
    font-size: 13px;
    color: var(--muted);
  }

  .toast {
    position: fixed;
    bottom: 24px;
    right: 24px;
    background: var(--danger);
    color: #fff;
    padding: 12px 20px;
    border-radius: 8px;
    font-size: 13px;
    font-family: var(--font-mono);
    transform: translateY(80px);
    opacity: 0;
    transition: all 0.3s;
    z-index: 999;
  }

  .toast.show { transform: translateY(0); opacity: 1; }

  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(8px); }
    to   { opacity: 1; transform: translateY(0); }
  }

  .ascii-out.visible { animation: fadeIn 0.4s ease; }

  ::-webkit-scrollbar { width: 6px; height: 6px; }
  ::-webkit-scrollbar-track { background: transparent; }
  ::-webkit-scrollbar-thumb { background: #222; border-radius: 3px; }
  ::-webkit-scrollbar-thumb:hover { background: #333; }
</style>
</head>
<body>
<div class="app">
  <header>
    <div class="logo">
      <span class="logo-text">goascii</span>
      <span class="logo-sub">image → ascii</span>
    </div>
    <a href="https://github.com/cubexteam/goascii" target="_blank" class="header-link">github ↗</a>
  </header>

  <main>
    <aside class="sidebar">

      <div>
        <div class="section-label">Image</div>
        <div class="dropzone" id="dropzone">
          <span class="dropzone-icon">⬡</span>
          <div class="dropzone-text">
            <strong>Drop image here</strong><br>
            or click to browse<br>
            <span style="font-size:11px">PNG · JPG · GIF</span>
          </div>
          <input type="file" id="fileInput" accept="image/png,image/jpeg,image/gif">
        </div>
        <img id="previewImg" class="preview-img" alt="preview">
      </div>

      <div class="control-group">
        <div class="section-label">Character Set</div>
        <div class="charset-grid">
          <button class="charset-btn active" data-cs="detailed">
            <span class="cs-name">detailed</span>
            <span class="cs-preview">$@B%8&WM#*o</span>
          </button>
          <button class="charset-btn" data-cs="simple">
            <span class="cs-name">simple</span>
            <span class="cs-preview">@#S%?*+;:,.</span>
          </button>
          <button class="charset-btn" data-cs="blocks">
            <span class="cs-name">blocks</span>
            <span class="cs-preview">█▓▒░ </span>
          </button>
          <button class="charset-btn" data-cs="braille">
            <span class="cs-name">braille</span>
            <span class="cs-preview">⣿⣷⣯⣟⡿⢿</span>
          </button>
        </div>
      </div>

      <div class="control-group">
        <div class="section-label">
          Width
        </div>
        <div class="slider-row">
          <span style="font-size:12px;color:var(--muted)">20</span>
          <span class="slider-val" id="widthVal">160</span>
          <span style="font-size:12px;color:var(--muted)">300</span>
        </div>
        <input type="range" id="widthSlider" min="20" max="300" value="160">
      </div>

      <div class="control-group">
        <div class="section-label">Options</div>
        <div class="toggle-row">
          <div>
            <span class="toggle-label">Color</span>
            <span class="toggle-desc">RGB per character</span>
          </div>
          <button class="toggle on" id="toggleColor" title="Color output"></button>
        </div>
        <div class="toggle-row">
          <div>
            <span class="toggle-label">Invert</span>
            <span class="toggle-desc">For light backgrounds</span>
          </div>
          <button class="toggle" id="toggleInvert" title="Invert brightness"></button>
        </div>
      </div>

      <button class="convert-btn" id="convertBtn" disabled>
        <span class="btn-text">Convert</span>
        <span class="btn-spinner"></span>
      </button>

    </aside>

    <div class="output-panel">
      <div class="output-toolbar">
        <div class="output-tabs">
          <button class="tab-btn active" data-tab="color">Color</button>
          <button class="tab-btn" data-tab="plain">Plain</button>
        </div>
        <div class="toolbar-spacer"></div>
        <button class="icon-btn" id="copyBtn">Copy</button>
        <button class="icon-btn accent" id="saveBtn">Save HTML</button>
      </div>

      <div class="output-scroll" id="outputScroll">
        <div class="empty-state" id="emptyState">
          <div class="empty-state-art">┌─────────────────────┐
│                     │
│   drop an image     │
│   to get started    │
│                     │
└─────────────────────┘</div>
          <div class="empty-state-text">by SantianDev</div>
        </div>
        <pre class="ascii-out" id="colorOut"></pre>
        <pre class="ascii-out" id="plainOut" style="color:#ccc"></pre>
      </div>
    </div>
  </main>
</div>

<div class="toast" id="toast"></div>

<script>
  let selectedFile = null;
  let activeCharset = 'detailed';
  let colorOn = true;
  let invertOn = false;
  let activeTab = 'color';
  let lastPlain = '';
  let lastHTML = '';

  const dropzone    = document.getElementById('dropzone');
  const fileInput   = document.getElementById('fileInput');
  const previewImg  = document.getElementById('previewImg');
  const widthSlider = document.getElementById('widthSlider');
  const widthVal    = document.getElementById('widthVal');
  const toggleColor  = document.getElementById('toggleColor');
  const toggleInvert = document.getElementById('toggleInvert');
  const convertBtn  = document.getElementById('convertBtn');
  const colorOut    = document.getElementById('colorOut');
  const plainOut    = document.getElementById('plainOut');
  const emptyState  = document.getElementById('emptyState');
  const copyBtn     = document.getElementById('copyBtn');
  const saveBtn     = document.getElementById('saveBtn');
  const toast       = document.getElementById('toast');

  dropzone.addEventListener('dragover', e => { e.preventDefault(); dropzone.classList.add('drag-over'); });
  dropzone.addEventListener('dragleave', () => dropzone.classList.remove('drag-over'));
  dropzone.addEventListener('drop', e => {
    e.preventDefault();
    dropzone.classList.remove('drag-over');
    const f = e.dataTransfer.files[0];
    if (f) setFile(f);
  });

  fileInput.addEventListener('change', () => {
    if (fileInput.files[0]) setFile(fileInput.files[0]);
  });

  function setFile(f) {
    selectedFile = f;
    const url = URL.createObjectURL(f);
    previewImg.src = url;
    previewImg.style.display = 'block';
    convertBtn.disabled = false;
  }

  document.querySelectorAll('.charset-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.charset-btn').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      activeCharset = btn.dataset.cs;
    });
  });

  widthSlider.addEventListener('input', () => {
    widthVal.textContent = widthSlider.value;
  });

  toggleColor.addEventListener('click', () => {
    colorOn = !colorOn;
    toggleColor.classList.toggle('on', colorOn);
  });

  toggleInvert.addEventListener('click', () => {
    invertOn = !invertOn;
    toggleInvert.classList.toggle('on', invertOn);
  });

  document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      activeTab = btn.dataset.tab;
      switchTab();
    });
  });

  function switchTab() {
    colorOut.classList.toggle('visible', activeTab === 'color');
    plainOut.classList.toggle('visible', activeTab === 'plain');
  }

  convertBtn.addEventListener('click', async () => {
    if (!selectedFile) return;

    convertBtn.classList.add('loading');
    convertBtn.disabled = true;

    const fd = new FormData();
    fd.append('image', selectedFile);
    fd.append('charset', activeCharset);
    fd.append('width', widthSlider.value);
    fd.append('invert', invertOn ? 'true' : 'false');
    fd.append('colored', colorOn ? 'true' : 'false');

    try {
      const res = await fetch('/convert', { method: 'POST', body: fd });
      const data = await res.json();

      if (data.error) {
        showToast(data.error);
        return;
      }

      lastHTML = data.html;
      lastPlain = data.plain;

      emptyState.style.display = 'none';

      colorOut.innerHTML = data.html;
      colorOut.classList.add('visible');

      plainOut.textContent = data.plain;
      plainOut.classList.add('visible');

      switchTab();

    } catch (e) {
      showToast('Connection error');
    } finally {
      convertBtn.classList.remove('loading');
      convertBtn.disabled = false;
    }
  });

  copyBtn.addEventListener('click', () => {
    if (!lastPlain) return;
    navigator.clipboard.writeText(lastPlain).then(() => showToast('Copied!', false));
  });

  saveBtn.addEventListener('click', () => {
    if (!lastHTML) return;
    const full = '<!DOCTYPE html><html><head><meta charset="UTF-8"><style>body{background:#0a0a0a;margin:20px}pre{font-family:"Courier New",monospace;font-size:8px;line-height:1.15}span{white-space:pre}</style></head><body><pre>' + lastHTML + '</pre></body></html>';
    const blob = new Blob([full], { type: 'text/html' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'ascii.html';
    a.click();
  });

  function showToast(msg, isError = true) {
    toast.textContent = msg;
    toast.style.background = isError ? 'var(--danger)' : 'var(--accent)';
    toast.style.color = isError ? '#fff' : '#000';
    toast.classList.add('show');
    setTimeout(() => toast.classList.remove('show'), 2500);
  }
</script>
</body>
</html>`
