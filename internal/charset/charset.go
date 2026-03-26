// Пакет charset содержит наборы символов для ASCII-арта
package charset

type Charset []rune

var (
	Detailed Charset = []rune(`$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\|()1{}[]?-_+~<>i!lI;:,"^` + "`.' ")
	Simple   Charset = []rune(`@#S%?*+;:,. `)
	Blocks   Charset = []rune(`█▓▒░ `)
	Braille  Charset = []rune(`⣿⣷⣯⣟⡿⢿⣻⣽⣼⣸⣰⣠⣀⡀ `)
)

func Get(name string) Charset {
	switch name {
	case "simple":
		return Simple
	case "blocks":
		return Blocks
	case "braille":
		return Braille
	default:
		return Detailed
	}
}

func Names() []string {
	return []string{"detailed", "simple", "blocks", "braille"}
}
