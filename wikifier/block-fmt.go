package wikifier

type fmtBlock struct {
	*parserBlock
}

func newFmtBlock(name string, b *parserBlock) block {
	return &fmtBlock{parserBlock: b}
}

func (b *fmtBlock) html(page *Page, el element) {
	el.setMeta("noIndent", true)
	el.setMeta("noTags", true)
	for _, item := range b.visiblePosContent() {
		// if it's a string, format it
		if str, ok := item.content.(string); ok {
			el.add(page.parseFormattedTextOpts(str, &formatterOptions{
				noEntities: true,
				pos:        item.position,
			}))
			continue
		}
		el.add(item.content)
	}
}
