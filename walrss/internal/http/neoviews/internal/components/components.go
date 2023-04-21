package components

func makeTitle(title string) string {
	if len(title) != 0 {
		title += " | "
	}
	title += "Walrss"
	return title
}

type MenuItem struct {
	Path string
	Text string
	Icon string
}
