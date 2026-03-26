# goascii

CLI tool to convert images into ASCII art with color support.

**Author:** SantianDev | [github.com/cubexteam](https://github.com/cubexteam)

## Quick Start

```bash
git clone https://github.com/cubexteam/goascii.git
cd goascii
```

**Windows:** double-click `run.bat`

**Linux/Mac:**
```bash
go build -o goascii ./cmd/goascii && ./goascii
```

Opens a web interface at `http://localhost:7821` automatically.

## Features

- Drag & drop image upload (PNG, JPG, GIF)
- Color ASCII art (RGB per character)
- Multiple character sets: `detailed`, `simple`, `blocks`, `braille`
- Adjustable width (20–300 chars)
- Brightness inversion for light backgrounds
- Save result as HTML file
- Copy plain text to clipboard
- Zero external dependencies — stdlib only

## Requirements

- Go 1.24+
- Any modern browser

## License

MIT
