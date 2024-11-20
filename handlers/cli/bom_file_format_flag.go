package cli

import (
	"fmt"
	"strings"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/alecthomas/kong"
)

var _ kong.MapperValue = (*BOMFileFormatFlag)(nil)

type BOMFileFormatFlag struct {
	Format cyclonedx.BOMFileFormat
}

// Decode implements kong.MapperValue.
func (b *BOMFileFormatFlag) Decode(ctx *kong.DecodeContext) error {
	var value string
	if err := ctx.Scan.PopValueInto("value", &value); err != nil {
		return err
	}

	switch strings.ToLower(value) {
	case "xml":
		b.Format = cyclonedx.BOMFileFormatXML
	case "json":
		b.Format = cyclonedx.BOMFileFormatJSON
	default:
		return fmt.Errorf("unknown BOM file format: %s", value)
	}

	return nil
}
