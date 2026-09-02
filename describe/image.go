package describe

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	corepb "github.com/projecteru2/core/rpc/gen"
)

func Images(msgs ...*corepb.ListImageMessage) {
	describeOr(msgs, describeImages)
}

func describeImages(msgs []*corepb.ListImageMessage) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Node", "Image", "Tags"})

	for _, msg := range msgs {
		for _, image := range msg.Images {
			rows := [][]string{
				{msg.Nodename},
				{image.Id},
				image.Tags,
			}
			t.AppendRows(toTableRows(rows))
		}
	}
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true, VAlign: text.VAlignMiddle},
		{Number: 2, AutoMerge: true},
	})
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Render()
}
